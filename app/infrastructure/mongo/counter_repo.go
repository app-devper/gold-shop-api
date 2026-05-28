package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CollectionCounters = "counters"

type CounterRepository struct {
	client *Client
}

func NewCounterRepository(client *Client) *CounterRepository {
	return &CounterRepository{client: client}
}

func (r *CounterRepository) NextSequence(ctx context.Context, key string) (int, error) {
	coll, err := r.client.CollectionFromCtx(ctx, CollectionCounters)
	if err != nil {
		return 0, err
	}

	filter := bson.M{"_id": key}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var result struct {
		Seq int `bson:"seq"`
	}
	err = coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return 0, fmt.Errorf("failed to get next sequence for %s: %w", key, err)
	}
	return result.Seq, nil
}
