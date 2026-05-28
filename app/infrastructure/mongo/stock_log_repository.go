package mongo

import (
	"context"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StockLogRepository struct {
	client *Client
}

func NewStockLogRepository(client *Client) *StockLogRepository {
	return &StockLogRepository{client: client}
}

func (r *StockLogRepository) coll(ctx context.Context) (*mongo.Collection, error) {
	return r.client.CollectionFromCtx(ctx, CollectionStockLogs)
}

func (r *StockLogRepository) Create(ctx context.Context, log *entity.StockLog) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	res, err := coll.InsertOne(ctx, log)
	if err != nil {
		return err
	}
	log.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *StockLogRepository) GetByProductID(ctx context.Context, productID primitive.ObjectID) ([]*entity.StockLog, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	cursor, err := coll.Find(ctx, bson.M{"product_id": productID}, options.Find().SetSort(bson.M{"created_at": -1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*entity.StockLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *StockLogRepository) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, limit, offset int) ([]*entity.StockLog, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := coll.Find(ctx, bson.M{"branch_id": branchID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*entity.StockLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}
