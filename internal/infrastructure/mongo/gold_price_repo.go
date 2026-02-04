package mongo

import (
	"context"
	"time"

	"github.com/devper-gold/gold-shop-api/internal/domain/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GoldPriceRepository implements repository.GoldPriceRepository
type GoldPriceRepository struct {
	collection *mongo.Collection
}

// NewGoldPriceRepository creates a new GoldPriceRepository
func NewGoldPriceRepository(client *Client) *GoldPriceRepository {
	return &GoldPriceRepository{
		collection: client.Collection(CollectionGoldPrices),
	}
}

// Create creates a new gold price record
func (r *GoldPriceRepository) Create(ctx context.Context, price *entity.GoldPrice) error {
	price.CreatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, price)
	if err != nil {
		return err
	}

	price.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetCurrent retrieves the current active gold price
func (r *GoldPriceRepository) GetCurrent(ctx context.Context) (*entity.GoldPrice, error) {
	var price entity.GoldPrice
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})

	err := r.collection.FindOne(ctx, bson.M{"is_active": true}, opts).Decode(&price)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &price, nil
}

// GetHistory retrieves gold price history
func (r *GoldPriceRepository) GetHistory(ctx context.Context, limit int) ([]*entity.GoldPrice, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var prices []*entity.GoldPrice
	if err := cursor.All(ctx, &prices); err != nil {
		return nil, err
	}
	return prices, nil
}

// GetByDateRange retrieves gold prices by date range
func (r *GoldPriceRepository) GetByDateRange(ctx context.Context, from, to string) ([]*entity.GoldPrice, error) {
	fromTime, _ := time.Parse("2006-01-02", from)
	toTime, _ := time.Parse("2006-01-02", to)
	toTime = toTime.Add(24 * time.Hour)

	filter := bson.M{
		"date": bson.M{
			"$gte": fromTime,
			"$lt":  toTime,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var prices []*entity.GoldPrice
	if err := cursor.All(ctx, &prices); err != nil {
		return nil, err
	}
	return prices, nil
}

// DeactivateAll deactivates all gold prices
func (r *GoldPriceRepository) DeactivateAll(ctx context.Context) error {
	_, err := r.collection.UpdateMany(ctx, bson.M{"is_active": true}, bson.M{"$set": bson.M{"is_active": false}})
	return err
}
