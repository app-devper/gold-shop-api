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

type rewardRepository struct {
	collection *mongo.Collection
}

func NewRewardRepository(client *Client) repository.RewardRepository {
	return &rewardRepository{
		collection: client.Collection("rewards"),
	}
}

func (r *rewardRepository) Create(ctx context.Context, reward *entity.Reward) error {
	reward.ID = primitive.NewObjectID()
	reward.CreatedAt = time.Now()
	reward.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, reward)
	return err
}

func (r *rewardRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Reward, error) {
	var reward entity.Reward
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&reward)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &reward, nil
}

func (r *rewardRepository) GetAll(ctx context.Context, activeOnly bool) ([]*entity.Reward, error) {
	filter := bson.M{}
	if activeOnly {
		filter["is_active"] = true
	}

	opts := options.Find().SetSort(bson.M{"point_cost": 1})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rewards []*entity.Reward
	if err := cursor.All(ctx, &rewards); err != nil {
		return nil, err
	}
	return rewards, nil
}

func (r *rewardRepository) Update(ctx context.Context, reward *entity.Reward) error {
	reward.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": reward.ID},
		bson.M{"$set": reward},
	)
	return err
}

func (r *rewardRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// RewardRedemptionRepository implementation
type rewardRedemptionRepository struct {
	collection *mongo.Collection
}

func NewRewardRedemptionRepository(client *Client) repository.RewardRedemptionRepository {
	return &rewardRedemptionRepository{
		collection: client.Collection("reward_redemptions"),
	}
}

func (r *rewardRedemptionRepository) Create(ctx context.Context, redemption *entity.RewardRedemption) error {
	redemption.ID = primitive.NewObjectID()
	redemption.RedeemedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, redemption)
	return err
}

func (r *rewardRedemptionRepository) GetByCustomerID(ctx context.Context, customerID primitive.ObjectID) ([]*entity.RewardRedemption, error) {
	filter := bson.M{"customer_id": customerID}
	opts := options.Find().SetSort(bson.M{"redeemed_at": -1}) // Newest first

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var redemptions []*entity.RewardRedemption
	if err := cursor.All(ctx, &redemptions); err != nil {
		return nil, err
	}
	return redemptions, nil
}

func (r *rewardRedemptionRepository) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, limit, offset int) ([]*entity.RewardRedemption, error) {
	filter := bson.M{"branch_id": branchID}
	opts := options.Find().
		SetSort(bson.M{"redeemed_at": -1}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var redemptions []*entity.RewardRedemption
	if err := cursor.All(ctx, &redemptions); err != nil {
		return nil, err
	}
	return redemptions, nil
}
