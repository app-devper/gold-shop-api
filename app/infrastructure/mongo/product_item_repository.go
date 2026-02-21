package mongo

import (
	"context"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProductItemRepository struct {
	collection *mongo.Collection
}

func NewProductItemRepository(client *Client) *ProductItemRepository {
	return &ProductItemRepository{
		collection: client.Collection(CollectionProductItems),
	}
}

func (r *ProductItemRepository) Create(ctx context.Context, item *entity.ProductItem) error {
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	res, err := r.collection.InsertOne(ctx, item)
	if err != nil {
		return err
	}
	item.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *ProductItemRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.ProductItem, error) {
	var item entity.ProductItem
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *ProductItemRepository) GetByBarcode(ctx context.Context, barcode string) (*entity.ProductItem, error) {
	var item entity.ProductItem
	err := r.collection.FindOne(ctx, bson.M{"barcode": barcode}).Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *ProductItemRepository) GetByProductID(ctx context.Context, productID primitive.ObjectID, status []entity.ProductStatus) ([]*entity.ProductItem, error) {
	filter := bson.M{"product_id": productID}
	if len(status) > 0 {
		filter["status"] = bson.M{"$in": status}
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []*entity.ProductItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ProductItemRepository) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.ProductStatus) ([]*entity.ProductItem, error) {
	filter := bson.M{"branch_id": branchID}
	if len(status) > 0 {
		filter["status"] = bson.M{"$in": status}
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []*entity.ProductItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ProductItemRepository) Update(ctx context.Context, item *entity.ProductItem) error {
	item.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": item.ID}, item)
	return err
}

func (r *ProductItemRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
