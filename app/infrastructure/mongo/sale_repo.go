package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// parseDateRange parses from/to date strings (YYYY-MM-DD) and returns a half-open [from, to) time range.
// If to is empty or invalid, it defaults to from + 1 day.
func parseDateRange(from, to string) (time.Time, time.Time, error) {
	fromTime, err := time.Parse("2006-01-02", from)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid 'from' date: %w", err)
	}
	toTime, toErr := time.Parse("2006-01-02", to)
	if toErr != nil || toTime.IsZero() {
		toTime = fromTime.Add(24 * time.Hour)
	} else {
		toTime = toTime.Add(24 * time.Hour)
	}
	return fromTime, toTime, nil
}

// SaleRepository implements repository.SaleRepository
type SaleRepository struct {
	client  *Client
	counter *CounterRepository
}

// NewSaleRepository creates a new SaleRepository
func NewSaleRepository(client *Client) *SaleRepository {
	return &SaleRepository{
		client:  client,
		counter: NewCounterRepository(client),
	}
}

func (r *SaleRepository) coll(ctx context.Context) (*mongo.Collection, error) {
	return r.client.CollectionFromCtx(ctx, CollectionSales)
}

// Create creates a new sale
func (r *SaleRepository) Create(ctx context.Context, sale *entity.Sale) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	sale.CreatedAt = time.Now()
	sale.UpdatedAt = time.Now()

	result, err := coll.InsertOne(ctx, sale)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return entity.ErrDuplicateKey
		}
		return err
	}

	sale.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID retrieves a sale by ID
func (r *SaleRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Sale, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var sale entity.Sale
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&sale)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &sale, nil
}

// GetBySaleNumber retrieves a sale by sale number
func (r *SaleRepository) GetBySaleNumber(ctx context.Context, saleNumber string) (*entity.Sale, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var sale entity.Sale
	err = coll.FindOne(ctx, bson.M{"sale_number": saleNumber}).Decode(&sale)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &sale, nil
}

// GetByBranchID retrieves sales by branch ID with filters
func (r *SaleRepository) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.SaleStatus, limit, offset int) ([]*entity.Sale, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"branch_id": branchID}
	if len(status) > 0 {
		filter["status"] = bson.M{"$in": status}
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sales []*entity.Sale
	if err := cursor.All(ctx, &sales); err != nil {
		return nil, err
	}
	return sales, nil
}

// GetByCustomerID retrieves sales by customer ID
func (r *SaleRepository) GetByCustomerID(ctx context.Context, customerID primitive.ObjectID, limit int) ([]*entity.Sale, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := coll.Find(ctx, bson.M{"customer_id": customerID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sales []*entity.Sale
	if err := cursor.All(ctx, &sales); err != nil {
		return nil, err
	}
	return sales, nil
}

// GetByDateRange retrieves sales by date range
func (r *SaleRepository) GetByDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) ([]*entity.Sale, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	fromTime, toTime, err := parseDateRange(from, to)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"branch_id": branchID,
		"created_at": bson.M{
			"$gte": fromTime,
			"$lt":  toTime,
		},
	}

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sales []*entity.Sale
	if err := cursor.All(ctx, &sales); err != nil {
		return nil, err
	}
	return sales, nil
}

// Update updates a sale
func (r *SaleRepository) Update(ctx context.Context, sale *entity.Sale) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	sale.UpdatedAt = time.Now()

	result, err := coll.ReplaceOne(ctx, bson.M{"_id": sale.ID}, sale)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return entity.ErrNotFound
	}
	return nil
}

// GenerateSaleNumber returns a unique sale number using an atomic per-day
// counter, formatted per SRS 6.2:
//
//	sell      → S{YYMMDD}{XXXX}    (e.g. S260503001)
//	buy_old   → B{YYMMDD}{XXXX}
//	exchange  → TR{YYMMDD}{XXXX}
//
// branchCode is folded into the counter key so two branches running concurrently
// don't collide on the same daily sequence, but it is intentionally NOT in the
// final number string (the SRS format omits it).
func (r *SaleRepository) GenerateSaleNumber(ctx context.Context, branchCode string, saleType entity.SaleType) (string, error) {
	now := time.Now()
	today := now.Format("20060102")
	yymmdd := now.Format("060102")

	prefix := saleNumberPrefix(saleType)
	counterKey := fmt.Sprintf("sale-%s-%s-%s", prefix, branchCode, today)

	nextNum, err := r.counter.NextSequence(ctx, counterKey)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s%04d", prefix, yymmdd, nextNum), nil
}

func saleNumberPrefix(t entity.SaleType) string {
	switch t {
	case entity.SaleTypeBuyOld:
		return "B"
	case entity.SaleTypeExchange:
		return "TR"
	default:
		return "S"
	}
}

// SumByBranchAndDateRange calculates total net sales for a branch and date range
func (r *SaleRepository) SumByBranchAndDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) (float64, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return 0, err
	}
	fromTime, toTime, err := parseDateRange(from, to)
	if err != nil {
		return 0, err
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"branch_id":  branchID,
			"created_at": bson.M{"$gte": fromTime, "$lt": toTime},
			"status":     entity.SaleStatusCompleted,
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$net_total"},
		}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Total float64 `bson:"total"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return 0, err
	}

	if len(results) == 0 {
		return 0, nil
	}
	return results[0].Total, nil
}

// CountByBranchAndDateRange counts sales for a branch and date range
func (r *SaleRepository) CountByBranchAndDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) (int64, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return 0, err
	}
	fromTime, toTime, err := parseDateRange(from, to)
	if err != nil {
		return 0, err
	}

	filter := bson.M{
		"branch_id":  branchID,
		"created_at": bson.M{"$gte": fromTime, "$lt": toTime},
		"status":     entity.SaleStatusCompleted,
	}

	return coll.CountDocuments(ctx, filter)
}

// SumCostByBranchAndDateRange calculates total cost of goods sold
func (r *SaleRepository) SumCostByBranchAndDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) (float64, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return 0, err
	}
	fromTime, toTime, err := parseDateRange(from, to)
	if err != nil {
		return 0, err
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"branch_id":  branchID,
			"created_at": bson.M{"$gte": fromTime, "$lt": toTime},
			"status":     entity.SaleStatusCompleted,
		}}},
		// Unwind items to sum costs
		{{Key: "$unwind", Value: "$items"}},
		{{Key: "$group", Value: bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$items.cost"}, // Note: Need to check if SaleItem has Cost field
		}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Total float64 `bson:"total"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return 0, err
	}

	if len(results) == 0 {
		return 0, nil
	}
	return results[0].Total, nil
}

// GetUnpaidByBranchID retrieves sales that are not fully paid
func (r *SaleRepository) GetUnpaidByBranchID(ctx context.Context, branchID primitive.ObjectID) ([]*entity.Sale, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"branch_id": branchID,
			"status":    bson.M{"$in": []entity.SaleStatus{entity.SaleStatusPending, entity.SaleStatusCompleted}},
		}}},
		{{Key: "$addFields", Value: bson.M{
			"total_paid": bson.M{"$sum": "$payments.amount"},
		}}},
		{{Key: "$match", Value: bson.M{
			"$expr": bson.M{"$gt": []interface{}{"$net_total", "$total_paid"}},
		}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sales []*entity.Sale
	if err := cursor.All(ctx, &sales); err != nil {
		return nil, err
	}
	return sales, nil
}

// GetTopSellingProducts retrieves top selling products based on quantity
func (r *SaleRepository) GetTopSellingProducts(ctx context.Context, branchID primitive.ObjectID, from, to string, limit int) ([]repository.TopProduct, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	fromTime, toTime, err := parseDateRange(from, to)
	if err != nil {
		return nil, err
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"branch_id":  branchID,
			"created_at": bson.M{"$gte": fromTime, "$lt": toTime},
			"status":     entity.SaleStatusCompleted,
		}}},
		{{Key: "$unwind", Value: "$items"}},
		{{Key: "$group", Value: bson.M{
			"_id":          "$items.product_id",
			"product_name": bson.M{"$first": "$items.product_name"},
			"total_qty":    bson.M{"$sum": 1},
			"total_rev":    bson.M{"$sum": "$items.total"},
		}}},
		{{Key: "$sort", Value: bson.M{"total_qty": -1}}},
		{{Key: "$limit", Value: int64(limit)}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []repository.TopProduct
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// GetEmployeePerformance retrieves sales performance per employee
func (r *SaleRepository) GetEmployeePerformance(ctx context.Context, branchID primitive.ObjectID, from, to string) ([]repository.EmployeePerformance, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	fromTime, toTime, err := parseDateRange(from, to)
	if err != nil {
		return nil, err
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"branch_id":  branchID,
			"created_at": bson.M{"$gte": fromTime, "$lt": toTime},
			"status":     entity.SaleStatusCompleted,
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":            "$user_id",
			"total_sales":    bson.M{"$sum": "$net_total"},
			"sale_count":     bson.M{"$sum": 1},
			"avg_sale_value": bson.M{"$avg": "$net_total"},
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "users",
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "user",
		}}},
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$user",
			"preserveNullAndEmptyArrays": true,
		}}},
		{{Key: "$project", Value: bson.M{
			"_id":            1,
			"full_name":      bson.M{"$ifNull": []interface{}{"$user.full_name", "Unknown"}},
			"total_sales":    1,
			"sale_count":     1,
			"avg_sale_value": 1,
		}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []repository.EmployeePerformance
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// GetSalesTrends retrieves daily sales trends
func (r *SaleRepository) GetSalesTrends(ctx context.Context, branchID primitive.ObjectID, from, to string) ([]repository.SalesTrend, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	fromTime, toTime, err := parseDateRange(from, to)
	if err != nil {
		return nil, err
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"branch_id":  branchID,
			"created_at": bson.M{"$gte": fromTime, "$lt": toTime},
			"status":     entity.SaleStatusCompleted,
		}}},
		{{Key: "$unwind", Value: "$items"}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{
					"format": "%Y-%m-%d",
					"date":   "$created_at",
				},
			},
			"revenue":    bson.M{"$sum": "$items.total"},
			"cost":       bson.M{"$sum": "$items.cost"},
			"sale_count": bson.M{"$addToSet": "$_id"},
		}}},
		{{Key: "$project", Value: bson.M{
			"_id":        1,
			"revenue":    1,
			"cost":       1,
			"profit":     bson.M{"$subtract": []interface{}{"$revenue", "$cost"}},
			"sale_count": bson.M{"$size": "$sale_count"},
		}}},
		{{Key: "$sort", Value: bson.M{"_id": 1}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []repository.SalesTrend
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
