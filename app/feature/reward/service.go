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
	txManager      repository.TransactionManager
}

func NewService(
	rewardRepo repository.RewardRepository,
	redemptionRepo repository.RewardRedemptionRepository,
	customerRepo repository.CustomerRepository,
	txManager repository.TransactionManager,
) *Service {
	return &Service{
		rewardRepo:     rewardRepo,
		redemptionRepo: redemptionRepo,
		customerRepo:   customerRepo,
		txManager:      txManager,
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
	customer, err := s.customerRepo.GetByID(ctx, customerID)
	if err != nil || customer == nil {
		return nil, errors.New("customer not found")
	}
	if !customer.IsMember || customer.Membership == nil {
		return nil, errors.New("customer is not a member")
	}

	reward, err := s.rewardRepo.GetByID(ctx, rewardID)
	if err != nil || reward == nil {
		return nil, errors.New("reward not found")
	}

	if !reward.IsValid() {
		return nil, errors.New("reward is not valid or out of stock")
	}

	if customer.Membership.Points < reward.PointsRequired {
		return nil, errors.New("insufficient points")
	}

	redemption := &entity.RewardRedemption{
		ID:          primitive.NewObjectID(),
		CustomerID:  customerID,
		RewardID:    rewardID,
		BranchID:    branchID,
		PointsUsed:  reward.PointsRequired,
		RedeemedAt:  time.Now(),
		ProcessedBy: processedBy,
	}

	txErr := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		customer.Membership.Points -= reward.PointsRequired
		customer.UpdatedAt = time.Now()
		if err := s.customerRepo.Update(txCtx, customer); err != nil {
			return err
		}

		reward.DeductQuantity()
		reward.UpdatedAt = time.Now()
		if err := s.rewardRepo.Update(txCtx, reward); err != nil {
			return err
		}

		return s.redemptionRepo.Create(txCtx, redemption)
	})
	if txErr != nil {
		return nil, txErr
	}

	return redemption, nil
}

func (s *Service) GetCustomerRedemptions(ctx context.Context, customerID primitive.ObjectID) ([]*entity.RewardRedemption, error) {
	return s.redemptionRepo.GetByCustomerID(ctx, customerID)
}
