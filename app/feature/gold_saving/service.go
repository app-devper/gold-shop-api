package gold_saving

import (
	"context"
	"errors"
	"fmt"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func unitLabel(t entity.GoldSavingType) string {
	if t == entity.GoldSavingByMoney {
		return "฿"
	}
	return "g"
}

// Service handles gold saving logic
type Service struct {
	goldSavingRepo repository.GoldSavingRepository
	goldPriceRepo  repository.GoldPriceRepository
	branchRepo     repository.BranchRepository
	customerRepo   repository.CustomerRepository
}

// NewService creates a new GoldSaving service
func NewService(
	goldSavingRepo repository.GoldSavingRepository,
	goldPriceRepo repository.GoldPriceRepository,
	branchRepo repository.BranchRepository,
	customerRepo repository.CustomerRepository,
) *Service {
	return &Service{
		goldSavingRepo: goldSavingRepo,
		goldPriceRepo:  goldPriceRepo,
		branchRepo:     branchRepo,
		customerRepo:   customerRepo,
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
	if savingType != entity.GoldSavingByMoney && savingType != entity.GoldSavingByWeight {
		return nil, errors.New("invalid saving type")
	}
	if minDeposit < 0 || minWithdrawal < 0 {
		return nil, errors.New("minimums must be non-negative")
	}

	branch, err := s.branchRepo.GetByID(ctx, branchID)
	if err != nil || branch == nil {
		return nil, errors.New("branch not found")
	}

	customer, err := s.customerRepo.GetByID(ctx, customerID)
	if err != nil || customer == nil {
		return nil, errors.New("customer not found")
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

// Deposit deposits to a gold saving account.
// `amount` unit follows account.SavingType: ByMoney = baht, ByWeight = grams.
func (s *Service) Deposit(ctx context.Context, accountID primitive.ObjectID, amount float64, userID primitive.ObjectID) (*entity.GoldSaving, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}

	account, err := s.goldSavingRepo.GetByID(ctx, accountID)
	if err != nil || account == nil {
		return nil, errors.New("account not found")
	}

	if account.Status != entity.GoldSavingStatusActive {
		return nil, errors.New("account is not active")
	}

	if account.MinDeposit > 0 && amount < account.MinDeposit {
		unit := unitLabel(account.SavingType)
		return nil, fmt.Errorf("amount is below minimum deposit (%g %s)", account.MinDeposit, unit)
	}

	// Get current gold price
	goldPrice, err := s.goldPriceRepo.GetCurrent(ctx)
	if err != nil || goldPrice == nil {
		return nil, errors.New("failed to get current gold price")
	}

	if err := account.Deposit(amount, goldPrice.GoldOrnamentSell, userID); err != nil {
		return nil, err
	}

	if err := s.goldSavingRepo.Update(ctx, account); err != nil {
		return nil, errors.New("failed to update account")
	}

	return account, nil
}

// Withdraw withdraws from a gold saving account.
// `amount` unit: asCash=true → baht; asCash=false → grams (physical gold).
func (s *Service) Withdraw(ctx context.Context, accountID primitive.ObjectID, amount float64, asCash bool, userID primitive.ObjectID) (*entity.GoldSaving, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}

	account, err := s.goldSavingRepo.GetByID(ctx, accountID)
	if err != nil || account == nil {
		return nil, errors.New("account not found")
	}

	if account.Status != entity.GoldSavingStatusActive {
		return nil, errors.New("account is not active")
	}

	if account.MinWithdrawal > 0 && amount < account.MinWithdrawal {
		// Withdraw min is denominated by withdraw mode, not account mode:
		// asCash=true → baht; asCash=false → grams.
		unit := "g"
		if asCash {
			unit = "฿"
		}
		return nil, fmt.Errorf("amount is below minimum withdrawal (%g %s)", account.MinWithdrawal, unit)
	}

	// Get current gold price
	goldPrice, err := s.goldPriceRepo.GetCurrent(ctx)
	if err != nil || goldPrice == nil {
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

	// Block close while either balance is non-trivial. RoundGram precision is 1e-6.
	if account.GoldBalance > 1e-6 || account.CashBalance > 0.01 {
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
