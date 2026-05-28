package mongo

import (
	"context"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserRepository implements repository.UserRepository
type UserRepository struct {
	client *Client
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(client *Client) *UserRepository {
	return &UserRepository{client: client}
}

func (r *UserRepository) coll(ctx context.Context) (*mongo.Collection, error) {
	return r.client.CollectionFromCtx(ctx, CollectionUsers)
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	result, err := coll.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return entity.ErrDuplicateKey
		}
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.User, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var user entity.User
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var user entity.User
	err = coll.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByBranchID retrieves users by branch ID
func (r *UserRepository) GetByBranchID(ctx context.Context, branchID primitive.ObjectID) ([]*entity.User, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	cursor, err := coll.Find(ctx, bson.M{"branch_id": branchID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*entity.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

// GetAll retrieves all users
func (r *UserRepository) GetAll(ctx context.Context) ([]*entity.User, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*entity.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	user.UpdatedAt = time.Now()

	result, err := coll.ReplaceOne(ctx, bson.M{"_id": user.ID}, user)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return entity.ErrNotFound
	}
	return nil
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
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
