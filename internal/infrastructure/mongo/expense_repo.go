package mongo

import (
	"context"
	"time"

	"github.com/devper-gold/gold-shop-api/internal/domain/entity"
	"github.com/devper-gold/gold-shop-api/internal/domain/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type expenseRepository struct {
	collection *mongo.Collection
}

func NewExpenseRepository(client *Client) repository.ExpenseRepository {
	return &expenseRepository{
		collection: client.Collection("expenses"),
	}
}

func (r *expenseRepository) Create(ctx context.Context, expense *entity.Expense) error {
	expense.ID = primitive.NewObjectID()
	expense.CreatedAt = time.Now()
	expense.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, expense)
	return err
}

func (r *expenseRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Expense, error) {
	var expense entity.Expense
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&expense)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &expense, nil
}

func (r *expenseRepository) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.ExpenseStatus, limit, offset int) ([]*entity.Expense, error) {
	filter := bson.M{"branch_id": branchID}
	if len(status) > 0 {
		filter["status"] = bson.M{"$in": status}
	}

	opts := options.Find().
		SetSort(bson.M{"date": -1}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
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
	fromDate, _ := time.Parse("2006-01-02", from)
	toDate, _ := time.Parse("2006-01-02", to)
	toDate = toDate.Add(24 * time.Hour) // Include end date

	filter := bson.M{
		"branch_id": branchID,
		"date": bson.M{
			"$gte": fromDate,
			"$lt":  toDate,
		},
	}

	opts := options.Find().SetSort(bson.M{"date": -1})
	cursor, err := r.collection.Find(ctx, filter, opts)
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
	filter := bson.M{"category_id": categoryID}
	cursor, err := r.collection.Find(ctx, filter)
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
	expense.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": expense.ID},
		bson.M{"$set": expense},
	)
	return err
}

func (r *expenseRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *expenseRepository) SumByBranchAndDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) (float64, error) {
	fromDate, _ := time.Parse("2006-01-02", from)
	toDate, _ := time.Parse("2006-01-02", to)
	toDate = toDate.Add(24 * time.Hour)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"branch_id": branchID,
				"date": bson.M{
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

	cursor, err := r.collection.Aggregate(ctx, pipeline)
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
