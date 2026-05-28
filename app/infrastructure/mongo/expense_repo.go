package mongo

import (
	"context"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type expenseRepository struct {
	client *Client
}

func NewExpenseRepository(client *Client) repository.ExpenseRepository {
	return &expenseRepository{client: client}
}

func (r *expenseRepository) coll(ctx context.Context) (*mongo.Collection, error) {
	return r.client.CollectionFromCtx(ctx, CollectionExpenses)
}

func (r *expenseRepository) Create(ctx context.Context, expense *entity.Expense) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	expense.ID = primitive.NewObjectID()
	expense.CreatedAt = time.Now()
	expense.UpdatedAt = time.Now()

	_, err = coll.InsertOne(ctx, expense)
	return err
}

func (r *expenseRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Expense, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var expense entity.Expense
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&expense)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &expense, nil
}

func (r *expenseRepository) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.ExpenseStatus, limit, offset int) ([]*entity.Expense, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"branch_id": branchID}
	if len(status) > 0 {
		filter["status"] = bson.M{"$in": status}
	}

	opts := options.Find().
		SetSort(bson.M{"expense_date": -1}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var expenses []*entity.Expense
	if err := cursor.All(ctx, &expenses); err != nil {
		return nil, err
	}
	return expenses, nil
}

func (r *expenseRepository) GetByDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) ([]*entity.Expense, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	fromDate, _ := time.Parse("2006-01-02", from)
	toDate, _ := time.Parse("2006-01-02", to)
	toDate = toDate.Add(24 * time.Hour) // Include end date

	filter := bson.M{
		"branch_id": branchID,
		"expense_date": bson.M{
			"$gte": fromDate,
			"$lt":  toDate,
		},
	}

	opts := options.Find().SetSort(bson.M{"expense_date": -1})
	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var expenses []*entity.Expense
	if err := cursor.All(ctx, &expenses); err != nil {
		return nil, err
	}
	return expenses, nil
}

func (r *expenseRepository) GetByCategoryID(ctx context.Context, categoryID primitive.ObjectID) ([]*entity.Expense, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"category_id": categoryID}
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var expenses []*entity.Expense
	if err := cursor.All(ctx, &expenses); err != nil {
		return nil, err
	}
	return expenses, nil
}

func (r *expenseRepository) Update(ctx context.Context, expense *entity.Expense) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	expense.UpdatedAt = time.Now()
	_, err = coll.UpdateOne(
		ctx,
		bson.M{"_id": expense.ID},
		bson.M{"$set": expense},
	)
	return err
}

func (r *expenseRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	_, err = coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *expenseRepository) SumByBranchAndDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) (float64, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return 0, err
	}
	fromDate, _ := time.Parse("2006-01-02", from)
	toDate, _ := time.Parse("2006-01-02", to)
	toDate = toDate.Add(24 * time.Hour)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"branch_id": branchID,
				"expense_date": bson.M{
					"$gte": fromDate,
					"$lt":  toDate,
				},
				"status": "approved", // Only sum approved expenses
			},
		},
		{
			"$group": bson.M{
				"_id":   nil,
				"total": bson.M{"$sum": "$amount"},
			},
		},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result struct {
			Total float64 `bson:"total"`
		}
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
		return result.Total, nil
	}

	return 0, nil
}
