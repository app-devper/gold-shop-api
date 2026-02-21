package gold_saving

import (
	"context"
	"errors"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service handles gold saving logic
type Service struct {
	goldSavingRepo repository.GoldSavingRepository
	goldPriceRepo  repository.GoldPriceRepository
	branchRepo     repository.BranchRepository
}

// NewService creates a new GoldSaving service
func NewService(
	goldSavingRepo repository.GoldSavingRepository,
	goldPriceRepo repository.GoldPriceRepository,
	branchRepo repository.BranchRepository,
) *Service {
	return &Service{
		goldSavingRepo: goldSavingRepo,
		goldPriceRepo:  goldPriceRepo,
		branchRepo:     branchRepo,
	}
}

// GetByID retrieves a gold saving account by ID
func (s *Service) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.GoldSaving, error) {
	return s.goldSavingRepo.GetByID(ctx, id)
}

// GetByBranchID retrieves gold saving accounts by branch
func (s *Service) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.GoldSavingStatus) ([]*entity.GoldSaving, error) {
	return s.goldSavingRepo.GetByBranchID(ctx, branchID, status)
}

// OpenAccount opens a new gold saving account
func (s *Service) OpenAccount(ctx context.Context, branchID, customerID primitive.ObjectID, savingType entity.GoldSavingType, minDeposit, minWithdrawal float64) (*entity.GoldSaving, error) {
	branch, err := s.branchRepo.GetByID(ctx, branchID)
	if err != nil || branch == nil {
		return nil, errors.New("branch not found")
	}

	accountNumber, err := s.goldSavingRepo.GenerateAccountNumber(ctx, branch.Code)
	if err != nil {
		return nil, errors.New("failed to generate account number")
	}

	account := entity.NewGoldSaving(branchID, customerID, accountNumber, savingType, minDeposit, minWithdrawal)

	if err := s.goldSavingRepo.Create(ctx, account); err != nil {
		return nil, errors.New("failed to create account")
	}

	return account, nil
}

// Deposit deposits to a gold saving account
func (s *Service) Deposit(ctx context.Context, accountID primitive.ObjectID, amount float64, userID primitive.ObjectID) (*entity.GoldSaving, error) {
	account, err := s.goldSavingRepo.GetByID(ctx, accountID)
	if err != nil || account == nil {
		return nil, errors.New("account not found")
	}

	if account.Status != entity.GoldSavingStatusActive {
		return nil, errors.New("account is not active")
	}

	if amount < account.MinDeposit {
		return nil, errors.New("amount is below minimum deposit")
	}

	// Get current gold price
	goldPrice, err := s.goldPriceRepo.GetCurrent(ctx)
	if err != nil {
		return nil, errors.New("failed to get current gold price")
	}

	if err := account.Deposit(amount, goldPrice.GoldBarSell, userID); err != nil {
		return nil, err
	}

	if err := s.goldSavingRepo.Update(ctx, account); err != nil {
		return nil, errors.New("failed to update account")
	}

	return account, nil
}

// Withdraw withdraws from a gold saving account
func (s *Service) Withdraw(ctx context.Context, accountID primitive.ObjectID, amount float64, asCash bool, userID primitive.ObjectID) (*entity.GoldSaving, error) {
	account, err := s.goldSavingRepo.GetByID(ctx, accountID)
	if err != nil || account == nil {
		return nil, errors.New("account not found")
	}

	if account.Status != entity.GoldSavingStatusActive {
		return nil, errors.New("account is not active")
	}

	if amount < account.MinWithdrawal {
		return nil, errors.New("amount is below minimum withdrawal")
	}

	// Get current gold price
	goldPrice, err := s.goldPriceRepo.GetCurrent(ctx)
	if err != nil {
		return nil, errors.New("failed to get current gold price")
	}

	if err := account.Withdraw(amount, asCash, goldPrice.GoldOrnamentBuy, userID); err != nil {
		return nil, err
	}

	if err := s.goldSavingRepo.Update(ctx, account); err != nil {
		return nil, errors.New("failed to update account")
	}

	return account, nil
}

// Close closes a gold saving account
func (s *Service) Close(ctx context.Context, accountID primitive.ObjectID) (*entity.GoldSaving, error) {
	account, err := s.goldSavingRepo.GetByID(ctx, accountID)
	if err != nil || account == nil {
		return nil, errors.New("account not found")
	}

	if account.Status != entity.GoldSavingStatusActive {
		return nil, errors.New("account is not active")
	}

	if account.GoldBalance > 0 {
		return nil, errors.New("account still has balance, please withdraw first")
	}

	account.Close()

	if err := s.goldSavingRepo.Update(ctx, account); err != nil {
		return nil, errors.New("failed to close account")
	}

	return account, nil
}

// Statement represents an account statement
type Statement struct {
	Account      *entity.GoldSaving `json:"account"`
	CurrentValue float64            `json:"current_value"`
	GoldPrice    float64            `json:"gold_price"`
}

// GetStatement returns account statement
func (s *Service) GetStatement(ctx context.Context, accountID primitive.ObjectID) (*Statement, error) {
	account, err := s.goldSavingRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	goldPrice, _ := s.goldPriceRepo.GetCurrent(ctx)
	currentPrice := 0.0
	if goldPrice != nil {
		currentPrice = goldPrice.GoldOrnamentBuy
	}

	return &Statement{
		Account:      account,
		CurrentValue: account.GetCurrentValue(currentPrice),
		GoldPrice:    currentPrice,
	}, nil
}
