package mongo

import (
	"context"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// BranchRepository implements repository.BranchRepository
type BranchRepository struct {
	client *Client
}

// NewBranchRepository creates a new BranchRepository
func NewBranchRepository(client *Client) *BranchRepository {
	return &BranchRepository{client: client}
}

func (r *BranchRepository) coll(ctx context.Context) (*mongo.Collection, error) {
	return r.client.CollectionFromCtx(ctx, CollectionBranches)
}

// Create creates a new branch
func (r *BranchRepository) Create(ctx context.Context, branch *entity.Branch) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	branch.CreatedAt = time.Now()
	branch.UpdatedAt = time.Now()

	result, err := coll.InsertOne(ctx, branch)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return entity.ErrDuplicateKey
		}
		return err
	}

	branch.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID retrieves a branch by ID
func (r *BranchRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Branch, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var branch entity.Branch
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&branch)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &branch, nil
}

// GetByCode retrieves a branch by code
func (r *BranchRepository) GetByCode(ctx context.Context, code string) (*entity.Branch, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var branch entity.Branch
	err = coll.FindOne(ctx, bson.M{"code": code}).Decode(&branch)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &branch, nil
}

// GetAll retrieves all branches
func (r *BranchRepository) GetAll(ctx context.Context) ([]*entity.Branch, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var branches []*entity.Branch
	if err := cursor.All(ctx, &branches); err != nil {
		return nil, err
	}
	return branches, nil
}

// Update updates a branch
func (r *BranchRepository) Update(ctx context.Context, branch *entity.Branch) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	branch.UpdatedAt = time.Now()

	result, err := coll.ReplaceOne(ctx, bson.M{"_id": branch.ID}, branch)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return entity.ErrNotFound
	}
	return nil
}

// Delete deletes a branch
func (r *BranchRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	result, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return entity.ErrNotFound
	}
	return nil
}
