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
	client *Client
}

func NewRewardRepository(client *Client) repository.RewardRepository {
	return &rewardRepository{client: client}
}

func (r *rewardRepository) coll(ctx context.Context) (*mongo.Collection, error) {
	return r.client.CollectionFromCtx(ctx, CollectionRewards)
}

func (r *rewardRepository) Create(ctx context.Context, reward *entity.Reward) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	reward.ID = primitive.NewObjectID()
	reward.CreatedAt = time.Now()
	reward.UpdatedAt = time.Now()

	_, err = coll.InsertOne(ctx, reward)
	return err
}

func (r *rewardRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Reward, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var reward entity.Reward
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&reward)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &reward, nil
}

func (r *rewardRepository) GetAll(ctx context.Context, activeOnly bool) ([]*entity.Reward, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	filter := bson.M{}
	if activeOnly {
		filter["is_active"] = true
	}

	opts := options.Find().SetSort(bson.M{"point_cost": 1})
	cursor, err := coll.Find(ctx, filter, opts)
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
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	reward.UpdatedAt = time.Now()
	_, err = coll.UpdateOne(
		ctx,
		bson.M{"_id": reward.ID},
		bson.M{"$set": reward},
	)
	return err
}

func (r *rewardRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	_, err = coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// RewardRedemptionRepository implementation
type rewardRedemptionRepository struct {
	client *Client
}

func NewRewardRedemptionRepository(client *Client) repository.RewardRedemptionRepository {
	return &rewardRedemptionRepository{client: client}
}

func (r *rewardRedemptionRepository) coll(ctx context.Context) (*mongo.Collection, error) {
	return r.client.CollectionFromCtx(ctx, CollectionRewardRedemptions)
}

func (r *rewardRedemptionRepository) Create(ctx context.Context, redemption *entity.RewardRedemption) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	redemption.ID = primitive.NewObjectID()
	redemption.RedeemedAt = time.Now()

	_, err = coll.InsertOne(ctx, redemption)
	return err
}

func (r *rewardRedemptionRepository) GetByCustomerID(ctx context.Context, customerID primitive.ObjectID) ([]*entity.RewardRedemption, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"customer_id": customerID}
	opts := options.Find().SetSort(bson.M{"redeemed_at": -1}) // Newest first

	cursor, err := coll.Find(ctx, filter, opts)
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
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"branch_id": branchID}
	opts := options.Find().
		SetSort(bson.M{"redeemed_at": -1}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := coll.Find(ctx, filter, opts)
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
