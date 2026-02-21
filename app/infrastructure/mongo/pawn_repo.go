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
}

// NewPawnRepository creates a new PawnRepository
func NewPawnRepository(client *Client) *PawnRepository {
	return &PawnRepository{
		collection: client.Collection(CollectionPawns),
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

// GeneratePawnNumber generates a unique pawn number
func (r *PawnRepository) GeneratePawnNumber(ctx context.Context, branchCode string) (string, error) {
	today := time.Now().Format("20060102")
	prefix := fmt.Sprintf("P-%s-%s-", branchCode, today)

	filter := bson.M{
		"pawn_number": bson.M{"$regex": "^" + prefix},
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "pawn_number", Value: -1}})

	var lastPawn entity.Pawn
	err := r.collection.FindOne(ctx, filter, opts).Decode(&lastPawn)

	var nextNum int
	if err == mongo.ErrNoDocuments {
		nextNum = 1
	} else if err != nil {
		return "", err
	} else {
		var lastNum int
		fmt.Sscanf(lastPawn.PawnNumber, prefix+"%04d", &lastNum)
		nextNum = lastNum + 1
	}

	return fmt.Sprintf("%s%04d", prefix, nextNum), nil
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
	fromTime, _ := time.Parse("2006-01-02", from)
	toTime, _ := time.Parse("2006-01-02", to)
	if !toTime.IsZero() {
		toTime = toTime.Add(24 * time.Hour)
	} else {
		toTime = fromTime.Add(24 * time.Hour)
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

	// Ejecuta redenciones
	cursorR, err := r.collection.Aggregate(ctx, redemptionPipeline)
	if err == nil {
		var resR []struct {
			Total float64 `bson:"total"`
		}
		if err := cursorR.All(ctx, &resR); err == nil && len(resR) > 0 {
			total += resR[0].Total
		}
		cursorR.Close(ctx)
	}

	// Ejecuta pagos de intereses
	cursorP, err := r.collection.Aggregate(ctx, paymentsPipeline)
	if err == nil {
		var resP []struct {
			Total float64 `bson:"total"`
		}
		if err := cursorP.All(ctx, &resP); err == nil && len(resP) > 0 {
			total += resP[0].Total
		}
		cursorP.Close(ctx)
	}

	return total, nil
}
