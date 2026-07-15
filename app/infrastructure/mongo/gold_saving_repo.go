package mongo

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GoldSavingRepository implements repository.GoldSavingRepository
type GoldSavingRepository struct {
	client *Client
}

// NewGoldSavingRepository creates a new GoldSavingRepository
func NewGoldSavingRepository(client *Client) *GoldSavingRepository {
	return &GoldSavingRepository{client: client}
}

func (r *GoldSavingRepository) coll(ctx context.Context) (*mongo.Collection, error) {
	return r.client.CollectionFromCtx(ctx, CollectionGoldSavings)
}

// Create creates a new gold saving account
func (r *GoldSavingRepository) Create(ctx context.Context, gs *entity.GoldSaving) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	gs.CreatedAt = time.Now()
	gs.UpdatedAt = time.Now()

	result, err := coll.InsertOne(ctx, gs)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return entity.ErrDuplicateKey
		}
		return err
	}

	gs.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID retrieves a gold saving account by ID
func (r *GoldSavingRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.GoldSaving, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var gs entity.GoldSaving
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&gs)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &gs, nil
}

// GetByAccountNumber retrieves by account number
func (r *GoldSavingRepository) GetByAccountNumber(ctx context.Context, accountNumber string) (*entity.GoldSaving, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var gs entity.GoldSaving
	err = coll.FindOne(ctx, bson.M{"account_number": accountNumber}).Decode(&gs)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &gs, nil
}

// GetByCustomerID retrieves by customer ID
func (r *GoldSavingRepository) GetByCustomerID(ctx context.Context, customerID primitive.ObjectID) ([]*entity.GoldSaving, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	cursor, err := coll.Find(ctx, bson.M{"customer_id": customerID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var accounts []*entity.GoldSaving
	if err := cursor.All(ctx, &accounts); err != nil {
		return nil, err
	}
	return accounts, nil
}

// GetByBranchID retrieves by branch ID
func (r *GoldSavingRepository) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.GoldSavingStatus) ([]*entity.GoldSaving, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"branch_id": branchID}
	if len(status) > 0 {
		filter["status"] = bson.M{"$in": status}
	}

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var accounts []*entity.GoldSaving
	if err := cursor.All(ctx, &accounts); err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *GoldSavingRepository) Search(ctx context.Context, branchID primitive.ObjectID, query string, customerIDs []primitive.ObjectID, limit int) ([]*entity.GoldSaving, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	matchers := []bson.M{{"account_number": bson.M{"$regex": regexp.QuoteMeta(query), "$options": "i"}}}
	if len(customerIDs) > 0 {
		matchers = append(matchers, bson.M{"customer_id": bson.M{"$in": customerIDs}})
	}
	filter := bson.M{"branch_id": branchID, "$or": matchers}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var accounts []*entity.GoldSaving
	if err := cursor.All(ctx, &accounts); err != nil {
		return nil, err
	}
	return accounts, nil
}

// Update updates a gold saving account
func (r *GoldSavingRepository) Update(ctx context.Context, gs *entity.GoldSaving) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	gs.UpdatedAt = time.Now()

	result, err := coll.ReplaceOne(ctx, bson.M{"_id": gs.ID}, gs)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return entity.ErrNotFound
	}
	return nil
}

// GenerateAccountNumber generates a unique account number
func (r *GoldSavingRepository) GenerateAccountNumber(ctx context.Context, branchCode string) (string, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return "", err
	}
	prefix := fmt.Sprintf("GS-%s-", branchCode)

	filter := bson.M{
		"account_number": bson.M{"$regex": "^" + prefix},
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "account_number", Value: -1}})

	var last entity.GoldSaving
	err = coll.FindOne(ctx, filter, opts).Decode(&last)

	var nextNum int
	if err == mongo.ErrNoDocuments {
		nextNum = 1
	} else if err != nil {
		return "", err
	} else {
		var lastNum int
		fmt.Sscanf(last.AccountNumber, prefix+"%06d", &lastNum)
		nextNum = lastNum + 1
	}

	return fmt.Sprintf("%s%06d", prefix, nextNum), nil
}

// SumBalanceByBranch calculates total cash balance of gold savings in a branch
func (r *GoldSavingRepository) SumBalanceByBranch(ctx context.Context, branchID primitive.ObjectID) (float64, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return 0, err
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"branch_id": branchID,
			"status":    entity.GoldSavingStatusActive,
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$cash_balance"},
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
