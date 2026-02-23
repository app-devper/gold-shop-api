package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PawnRepository implements repository.PawnRepository
type PawnRepository struct {
	collection *mongo.Collection
	counter    *CounterRepository
}

// NewPawnRepository creates a new PawnRepository
func NewPawnRepository(client *Client) *PawnRepository {
	return &PawnRepository{
		collection: client.Collection(CollectionPawns),
		counter:    NewCounterRepository(client),
	}
}

// Create creates a new pawn
func (r *PawnRepository) Create(ctx context.Context, pawn *entity.Pawn) error {
	pawn.CreatedAt = time.Now()
	pawn.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, pawn)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return entity.ErrDuplicateKey
		}
		return err
	}

	pawn.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID retrieves a pawn by ID
func (r *PawnRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Pawn, error) {
	var pawn entity.Pawn
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&pawn)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &pawn, nil
}

// GetByPawnNumber retrieves a pawn by pawn number
func (r *PawnRepository) GetByPawnNumber(ctx context.Context, pawnNumber string) (*entity.Pawn, error) {
	var pawn entity.Pawn
	err := r.collection.FindOne(ctx, bson.M{"pawn_number": pawnNumber}).Decode(&pawn)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &pawn, nil
}

// GetByBranchID retrieves pawns by branch ID
func (r *PawnRepository) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.PawnStatus, limit, offset int) ([]*entity.Pawn, error) {
	filter := bson.M{"branch_id": branchID}
	if len(status) > 0 {
		filter["status"] = bson.M{"$in": status}
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var pawns []*entity.Pawn
	if err := cursor.All(ctx, &pawns); err != nil {
		return nil, err
	}
	return pawns, nil
}

// GetByCustomerID retrieves pawns by customer ID
func (r *PawnRepository) GetByCustomerID(ctx context.Context, customerID primitive.ObjectID) ([]*entity.Pawn, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"customer_id": customerID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var pawns []*entity.Pawn
	if err := cursor.All(ctx, &pawns); err != nil {
		return nil, err
	}
	return pawns, nil
}

// GetDueSoon retrieves pawns due within specified days
func (r *PawnRepository) GetDueSoon(ctx context.Context, branchID primitive.ObjectID, days int) ([]*entity.Pawn, error) {
	now := time.Now()
	dueDate := now.AddDate(0, 0, days)

	filter := bson.M{
		"branch_id": branchID,
		"status":    entity.PawnStatusActive,
		"due_date": bson.M{
			"$gte": now,
			"$lte": dueDate,
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var pawns []*entity.Pawn
	if err := cursor.All(ctx, &pawns); err != nil {
		return nil, err
	}
	return pawns, nil
}

// GetOverdue retrieves overdue pawns
func (r *PawnRepository) GetOverdue(ctx context.Context, branchID primitive.ObjectID) ([]*entity.Pawn, error) {
	filter := bson.M{
		"branch_id": branchID,
		"status":    entity.PawnStatusActive,
		"due_date":  bson.M{"$lt": time.Now()},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var pawns []*entity.Pawn
	if err := cursor.All(ctx, &pawns); err != nil {
		return nil, err
	}
	return pawns, nil
}

// Update updates a pawn
func (r *PawnRepository) Update(ctx context.Context, pawn *entity.Pawn) error {
	pawn.UpdatedAt = time.Now()

	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": pawn.ID}, pawn)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return entity.ErrNotFound
	}
	return nil
}

// GeneratePawnNumber generates a unique pawn number using atomic counter
func (r *PawnRepository) GeneratePawnNumber(ctx context.Context, branchCode string) (string, error) {
	today := time.Now().Format("20060102")
	counterKey := fmt.Sprintf("pawn-%s-%s", branchCode, today)

	nextNum, err := r.counter.NextSequence(ctx, counterKey)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("P-%s-%s-%04d", branchCode, today, nextNum), nil
}

// CountByStatus counts pawns by status
func (r *PawnRepository) CountByStatus(ctx context.Context, branchID primitive.ObjectID, status entity.PawnStatus) (int64, error) {
	filter := bson.M{
		"branch_id": branchID,
		"status":    status,
	}
	return r.collection.CountDocuments(ctx, filter)
}

// SumInterestByBranchAndDateRange calculates total interest collected
func (r *PawnRepository) SumInterestByBranchAndDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) (float64, error) {
	fromTime, toTime, err := parseDateRange(from, to)
	if err != nil {
		return 0, err
	}

	// Sum interest from redemptions
	redemptionPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"branch_id":       branchID,
			"redemption.date": bson.M{"$gte": fromTime, "$lt": toTime},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$redemption.interest"},
		}}},
	}

	// Sum interest from interest payments
	paymentsPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"branch_id": branchID}}},
		{{Key: "$unwind", Value: "$interest_payments"}},
		{{Key: "$match", Value: bson.M{
			"interest_payments.payment_date": bson.M{"$gte": fromTime, "$lt": toTime},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$interest_payments.amount"},
		}}},
	}

	var total float64

	cursorR, err := r.collection.Aggregate(ctx, redemptionPipeline)
	if err != nil {
		return 0, fmt.Errorf("failed to aggregate redemption interest: %w", err)
	}
	var resR []struct {
		Total float64 `bson:"total"`
	}
	if err := cursorR.All(ctx, &resR); err != nil {
		cursorR.Close(ctx)
		return 0, fmt.Errorf("failed to decode redemption interest: %w", err)
	}
	cursorR.Close(ctx)
	if len(resR) > 0 {
		total += resR[0].Total
	}

	cursorP, err := r.collection.Aggregate(ctx, paymentsPipeline)
	if err != nil {
		return 0, fmt.Errorf("failed to aggregate interest payments: %w", err)
	}
	var resP []struct {
		Total float64 `bson:"total"`
	}
	if err := cursorP.All(ctx, &resP); err != nil {
		cursorP.Close(ctx)
		return 0, fmt.Errorf("failed to decode interest payments: %w", err)
	}
	cursorP.Close(ctx)
	if len(resP) > 0 {
		total += resP[0].Total
	}

	return total, nil
}
