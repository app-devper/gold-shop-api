package mongo

import (
	"context"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ProductRepository implements repository.ProductRepository.
type ProductRepository struct {
	collection *mongo.Collection
}

func NewProductRepository(client *Client) *ProductRepository {
	return &ProductRepository{collection: client.Collection(CollectionProducts)}
}

func (r *ProductRepository) Create(ctx context.Context, product *entity.Product) error {
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, product)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return entity.ErrDuplicateKey
		}
		return err
	}
	product.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Product, error) {
	var product entity.Product
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&product); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) GetBySKU(ctx context.Context, sku string) (*entity.Product, error) {
	var product entity.Product
	if err := r.collection.FindOne(ctx, bson.M{"sku": sku}).Decode(&product); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &product, nil
}

// GetByBranchID lists catalog masters by branch.
// Optional filters:
//
//	kind   — narrow to ornament|bar
//	search — case-insensitive regex against name + sku + design
func (r *ProductRepository) GetByBranchID(
	ctx context.Context, branchID primitive.ObjectID, kind entity.ProductKind, search string, limit, offset int,
) ([]*entity.Product, error) {
	filter := bson.M{"branch_id": branchID, "is_active": true}
	if kind != "" {
		filter["kind"] = kind
	}
	if search != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": search, "$options": "i"}},
			{"sku": bson.M{"$regex": search, "$options": "i"}},
			{"design": bson.M{"$regex": search, "$options": "i"}},
		}
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

	var products []*entity.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepository) Update(ctx context.Context, product *entity.Product) error {
	product.UpdatedAt = time.Now()
	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": product.ID}, product)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return entity.ErrNotFound
	}
	return nil
}

// Delete soft-deletes by flipping IsActive=false; we keep the row so historical
// sales can still resolve product_id → name.
func (r *ProductRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"is_active": false, "updated_at": time.Now()}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *ProductRepository) Count(ctx context.Context, branchID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"branch_id": branchID, "is_active": true})
}
