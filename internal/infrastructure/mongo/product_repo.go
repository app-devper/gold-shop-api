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

// ProductRepository implements repository.ProductRepository
type ProductRepository struct {
	collection *mongo.Collection
}

// NewProductRepository creates a new ProductRepository
func NewProductRepository(client *Client) *ProductRepository {
	return &ProductRepository{
		collection: client.Collection(CollectionProducts),
	}
}

// Create creates a new product
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

// GetByID retrieves a product by ID
func (r *ProductRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Product, error) {
	var product entity.Product
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &product, nil
}

// GetBySKU retrieves a product by SKU
func (r *ProductRepository) GetBySKU(ctx context.Context, sku string) (*entity.Product, error) {
	var product entity.Product
	err := r.collection.FindOne(ctx, bson.M{"sku": sku}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &product, nil
}

// GetByBarcode retrieves a product by barcode
func (r *ProductRepository) GetByBarcode(ctx context.Context, barcode string) (*entity.Product, error) {
	var product entity.Product
	err := r.collection.FindOne(ctx, bson.M{"barcode": barcode}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &product, nil
}

// GetByBranchID retrieves products by branch ID
func (r *ProductRepository) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.ProductStatus, limit, offset int) ([]*entity.Product, error) {
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

	var products []*entity.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}
	return products, nil
}

// GetByCategoryID retrieves products by category ID
func (r *ProductRepository) GetByCategoryID(ctx context.Context, categoryID primitive.ObjectID) ([]*entity.Product, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"category_id": categoryID})
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

// Search searches products by name, SKU, or barcode
func (r *ProductRepository) Search(ctx context.Context, branchID primitive.ObjectID, query string, limit int) ([]*entity.Product, error) {
	filter := bson.M{
		"branch_id": branchID,
		"$or": []bson.M{
			{"name": bson.M{"$regex": query, "$options": "i"}},
			{"sku": bson.M{"$regex": query, "$options": "i"}},
			{"barcode": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	opts := options.Find().SetLimit(int64(limit))
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

// GetLowStock retrieves products below reorder point
func (r *ProductRepository) GetLowStock(ctx context.Context, branchID primitive.ObjectID) ([]*entity.Product, error) {
	filter := bson.M{
		"branch_id":     branchID,
		"status":        entity.ProductStatusAvailable,
		"reorder_point": bson.M{"$gt": 0},
	}

	cursor, err := r.collection.Find(ctx, filter)
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

// Update updates a product
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

// Delete deletes a product
func (r *ProductRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return entity.ErrNotFound
	}
	return nil
}

// Count returns total product count for a branch
func (r *ProductRepository) Count(ctx context.Context, branchID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"branch_id": branchID})
}
