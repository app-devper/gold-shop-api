package reward

import (
	"context"
	"errors"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	rewardRepo     repository.RewardRepository
	redemptionRepo repository.RewardRedemptionRepository
	customerRepo   repository.CustomerRepository
}

func NewService(
	rewardRepo repository.RewardRepository,
	redemptionRepo repository.RewardRedemptionRepository,
	customerRepo repository.CustomerRepository,
) *Service {
	return &Service{
		rewardRepo:     rewardRepo,
		redemptionRepo: redemptionRepo,
		customerRepo:   customerRepo,
	}
}

func (s *Service) GetRewards(ctx context.Context, onlyActive bool) ([]*entity.Reward, error) {
	return s.rewardRepo.GetAll(ctx, onlyActive)
}

func (s *Service) GetReward(ctx context.Context, id primitive.ObjectID) (*entity.Reward, error) {
	return s.rewardRepo.GetByID(ctx, id)
}

func (s *Service) CreateReward(ctx context.Context, reward *entity.Reward) error {
	return s.rewardRepo.Create(ctx, reward)
}

func (s *Service) UpdateReward(ctx context.Context, reward *entity.Reward) error {
	reward.UpdatedAt = time.Now()
	return s.rewardRepo.Update(ctx, reward)
}

func (s *Service) RedeemReward(ctx context.Context, customerID, rewardID, branchID, processedBy primitive.ObjectID) (*entity.RewardRedemption, error) {
	// Get customer
	customer, err := s.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return nil, errors.New("customer not found")
	}

	// Get reward
	reward, err := s.rewardRepo.GetByID(ctx, rewardID)
	if err != nil {
		return nil, errors.New("reward not found")
	}

	// Validate reward
	if reward == nil || !reward.IsValid() {
		return nil, errors.New("reward is not valid or out of stock")
	}

	// Validate points
	if customer.Membership.Points < reward.PointsRequired {
		return nil, errors.New("insufficient points")
	}

	// Create redemption
	redemption := &entity.RewardRedemption{
		ID:          primitive.NewObjectID(),
		CustomerID:  customerID,
		RewardID:    rewardID,
		BranchID:    branchID,
		PointsUsed:  reward.PointsRequired,
		RedeemedAt:  time.Now(),
		ProcessedBy: processedBy,
	}

	// Deduct points from customer
	customer.Membership.Points -= reward.PointsRequired
	customer.UpdatedAt = time.Now()
	if err := s.customerRepo.Update(ctx, customer); err != nil {
		return nil, err
	}

	// Deduct quantity from reward
	reward.DeductQuantity()
	reward.UpdatedAt = time.Now()
	if err := s.rewardRepo.Update(ctx, reward); err != nil {
		return nil, err
	}

	// Save redemption
	if err := s.redemptionRepo.Create(ctx, redemption); err != nil {
		return nil, err
	}

	return redemption, nil
}

func (s *Service) GetCustomerRedemptions(ctx context.Context, customerID primitive.ObjectID) ([]*entity.RewardRedemption, error) {
	return s.redemptionRepo.GetByCustomerID(ctx, customerID)
}
