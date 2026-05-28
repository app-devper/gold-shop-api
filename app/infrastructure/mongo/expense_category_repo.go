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

type expenseCategoryRepository struct {
	client *Client
}

func NewExpenseCategoryRepository(client *Client) repository.ExpenseCategoryRepository {
	return &expenseCategoryRepository{client: client}
}

func (r *expenseCategoryRepository) coll(ctx context.Context) (*mongo.Collection, error) {
	return r.client.CollectionFromCtx(ctx, CollectionExpenseCategories)
}

func (r *expenseCategoryRepository) Create(ctx context.Context, category *entity.ExpenseCategory) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	category.ID = primitive.NewObjectID()
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()

	_, err = coll.InsertOne(ctx, category)
	return err
}

func (r *expenseCategoryRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.ExpenseCategory, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var category entity.ExpenseCategory
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *expenseCategoryRepository) GetAll(ctx context.Context) ([]*entity.ExpenseCategory, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var categories []*entity.ExpenseCategory
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *expenseCategoryRepository) Update(ctx context.Context, category *entity.ExpenseCategory) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	category.UpdatedAt = time.Now()
	_, err = coll.UpdateOne(
		ctx,
		bson.M{"_id": category.ID},
		bson.M{"$set": category},
	)
	return err
}

func (r *expenseCategoryRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	_, err = coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
