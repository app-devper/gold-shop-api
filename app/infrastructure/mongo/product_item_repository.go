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
	client *Client
}

func NewProductItemRepository(client *Client) *ProductItemRepository {
	return &ProductItemRepository{client: client}
}

func (r *ProductItemRepository) coll(ctx context.Context) (*mongo.Collection, error) {
	return r.client.CollectionFromCtx(ctx, CollectionProductItems)
}

func (r *ProductItemRepository) Create(ctx context.Context, item *entity.ProductItem) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	res, err := coll.InsertOne(ctx, item)
	if err != nil {
		return err
	}
	item.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *ProductItemRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.ProductItem, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var item entity.ProductItem
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *ProductItemRepository) GetByBarcode(ctx context.Context, barcode string) (*entity.ProductItem, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var item entity.ProductItem
	err = coll.FindOne(ctx, bson.M{"barcode": barcode}).Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *ProductItemRepository) GetByProductID(ctx context.Context, productID primitive.ObjectID, status []entity.ProductStatus) ([]*entity.ProductItem, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"product_id": productID}
	if len(status) > 0 {
		filter["status"] = bson.M{"$in": status}
	}

	cursor, err := coll.Find(ctx, filter)
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

	var items []*entity.ProductItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ProductItemRepository) Update(ctx context.Context, item *entity.ProductItem) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	item.UpdatedAt = time.Now()
	_, err = coll.ReplaceOne(ctx, bson.M{"_id": item.ID}, item)
	return err
}

func (r *ProductItemRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	_, err = coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
