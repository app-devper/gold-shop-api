package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CollectionCounters = "counters"

// CounterRepository provides atomic counter operations
type CounterRepository struct {
	collection *mongo.Collection
}

// NewCounterRepository creates a new CounterRepository
func NewCounterRepository(client *Client) *CounterRepository {
	return &CounterRepository{
		collection: client.Collection(CollectionCounters),
	}
}

// NextSequence atomically increments and returns the next sequence number for a given key
func (r *CounterRepository) NextSequence(ctx context.Context, key string) (int, error) {
	filter := bson.M{"_id": key}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var result struct {
		Seq int `bson:"seq"`
	}
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return 0, fmt.Errorf("failed to get next sequence for %s: %w", key, err)
	}
	return result.Seq, nil
}
