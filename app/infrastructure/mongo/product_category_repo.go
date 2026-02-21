package mongo

import (
	"context"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type productCategoryRepository struct {
	collection *mongo.Collection
}

func NewProductCategoryRepository(client *Client) repository.ProductCategoryRepository {
	return &productCategoryRepository{
		collection: client.Collection("product_categories"),
	}
}

func (r *productCategoryRepository) Create(ctx context.Context, category *entity.ProductCategory) error {
	category.ID = primitive.NewObjectID()
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, category)
	return err
}

func (r *productCategoryRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.ProductCategory, error) {
	var category entity.ProductCategory
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *productCategoryRepository) GetAll(ctx context.Context) ([]*entity.ProductCategory, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var categories []*entity.ProductCategory
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *productCategoryRepository) Update(ctx context.Context, category *entity.ProductCategory) error {
	category.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": category.ID},
		bson.M{"$set": category},
	)
	return err
}

func (r *productCategoryRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
