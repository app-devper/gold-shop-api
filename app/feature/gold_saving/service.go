package gold_saving

import (
	"context"
	"errors"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BahtPerGramOrnament is the ornament-gold conversion ratio.
// Per-gram price = price-per-baht / BahtPerGramOrnament.
const BahtPerGramOrnament = entity.BahtPerGramOrnament

// Service implements the unified single-balance gold-savings business logic.
type Service struct {
	goldSavingRepo repository.GoldSavingRepository
	goldPriceRepo  repository.GoldPriceRepository
	branchRepo     repository.BranchRepository
	customerRepo   repository.CustomerRepository
}

// NewService creates a new GoldSaving service.
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

// OpenInput captures account-creation parameters.
// Min thresholds are optional; pass 0 to disable.
type OpenInput struct {
	BranchID            primitive.ObjectID
	CustomerID          primitive.ObjectID
	MinDepositCash      float64
	MinDepositPhysical  float64
	MinWithdrawCash     float64
	MinWithdrawPhysical float64
}

// Open creates a fresh active account for the customer.
func (s *Service) Open(ctx context.Context, in OpenInput) (*entity.GoldSaving, error) {
	branch, err := s.branchRepo.GetByID(ctx, in.BranchID)
	if err != nil || branch == nil {
		return nil, errors.New("branch not found")
	}
	customer, err := s.customerRepo.GetByID(ctx, in.CustomerID)
	if err != nil || customer == nil {
		return nil, errors.New("customer not found")
	}
	accountNumber, err := s.goldSavingRepo.GenerateAccountNumber(ctx, branch.Code)
	if err != nil {
		return nil, errors.New("failed to generate account number")
	}

	account := entity.NewGoldSaving(in.BranchID, in.CustomerID, accountNumber)
	account.SetMinimums(in.MinDepositCash, in.MinDepositPhysical, in.MinWithdrawCash, in.MinWithdrawPhysical)

	if err := s.goldSavingRepo.Create(ctx, account); err != nil {
		return nil, errors.New("failed to create account")
	}
	return account, nil
}

// GetByID retrieves an account.
func (s *Service) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.GoldSaving, error) {
	return s.goldSavingRepo.GetByID(ctx, id)
}

// GetByBranchID lists accounts in a branch optionally filtered by status.
func (s *Service) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.GoldSavingStatus) ([]*entity.GoldSaving, error) {
	return s.goldSavingRepo.GetByBranchID(ctx, branchID, status)
}

// DepositCash converts a baht amount into gold weight at the current sell price.
func (s *Service) DepositCash(ctx context.Context, accountID primitive.ObjectID, amountBaht float64, by primitive.ObjectID) (*entity.GoldSaving, error) {
	account, sellPerGram, _, err := s.loadActiveAccountWithPrices(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if err := account.DepositCash(amountBaht, sellPerGram, by); err != nil {
		return nil, err
	}
	if err := s.goldSavingRepo.Update(ctx, account); err != nil {
		return nil, errors.New("failed to update account")
	}
	return account, nil
}

// DepositGold credits a physical-gold deposit at the operator-stated weight.
func (s *Service) DepositGold(ctx context.Context, accountID primitive.ObjectID, weight float64, by primitive.ObjectID) (*entity.GoldSaving, error) {
	account, sellPerGram, _, err := s.loadActiveAccountWithPrices(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if err := account.DepositGold(weight, sellPerGram, by); err != nil {
		return nil, err
	}
	if err := s.goldSavingRepo.Update(ctx, account); err != nil {
		return nil, errors.New("failed to update account")
	}
	return account, nil
}

// WithdrawCash converts a baht cashout into a gold-weight debit at the current buy price.
func (s *Service) WithdrawCash(ctx context.Context, accountID primitive.ObjectID, amountBaht float64, by primitive.ObjectID) (*entity.GoldSaving, error) {
	account, _, buyPerGram, err := s.loadActiveAccountWithPrices(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if err := account.WithdrawCash(amountBaht, buyPerGram, by); err != nil {
		return nil, err
	}
	if err := s.goldSavingRepo.Update(ctx, account); err != nil {
		return nil, errors.New("failed to update account")
	}
	return account, nil
}

// WithdrawGold debits a physical-gold withdrawal at the operator-stated weight.
func (s *Service) WithdrawGold(ctx context.Context, accountID primitive.ObjectID, weight float64, by primitive.ObjectID) (*entity.GoldSaving, error) {
	account, _, buyPerGram, err := s.loadActiveAccountWithPrices(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if err := account.WithdrawGold(weight, buyPerGram, by); err != nil {
		return nil, err
	}
	if err := s.goldSavingRepo.Update(ctx, account); err != nil {
		return nil, errors.New("failed to update account")
	}
	return account, nil
}

// AdjustInput is an admin manual correction.
// WeightDelta is signed (grams). Note is required.
type AdjustInput struct {
	AccountID   primitive.ObjectID
	WeightDelta float64
	Note        string
	By          primitive.ObjectID
}

// Adjust applies an admin correction; both note and a non-zero delta are required.
// The current buy price (if available) is captured on the audit entry as the cash equivalent.
func (s *Service) Adjust(ctx context.Context, in AdjustInput) (*entity.GoldSaving, error) {
	if in.Note == "" {
		return nil, errors.New("note is required for adjustment")
	}
	account, err := s.goldSavingRepo.GetByID(ctx, in.AccountID)
	if err != nil || account == nil {
		return nil, errors.New("account not found")
	}
	if account.Status != entity.GoldSavingStatusActive {
		return nil, errors.New("account is not active")
	}
	priceForAudit := 0.0
	if gp, _ := s.goldPriceRepo.GetCurrent(ctx); gp != nil {
		priceForAudit = gp.GoldOrnamentBuy / BahtPerGramOrnament
	}
	if err := account.Adjust(in.WeightDelta, priceForAudit, in.Note, in.By); err != nil {
		return nil, err
	}
	if err := s.goldSavingRepo.Update(ctx, account); err != nil {
		return nil, errors.New("failed to update account")
	}
	return account, nil
}

// Close closes an active account. Balance must be effectively zero (≤ 1e-6 g).
func (s *Service) Close(ctx context.Context, accountID primitive.ObjectID) (*entity.GoldSaving, error) {
	account, err := s.goldSavingRepo.GetByID(ctx, accountID)
	if err != nil || account == nil {
		return nil, errors.New("account not found")
	}
	if account.Status != entity.GoldSavingStatusActive {
		return nil, errors.New("account is not active")
	}
	if account.GoldWeight > 1e-6 {
		return nil, errors.New("account still has balance, please withdraw first")
	}
	account.Close()
	if err := s.goldSavingRepo.Update(ctx, account); err != nil {
		return nil, errors.New("failed to close account")
	}
	return account, nil
}

// Statement is the mark-to-market view for the customer.
type Statement struct {
	Account             *entity.GoldSaving `json:"account"`
	GoldWeight          float64            `json:"gold_weight"`            // grams
	CurrentBuyPrice     float64            `json:"current_buy_price"`      // ฿ per baht (display)
	CurrentBuyPerGram   float64            `json:"current_buy_per_gram"`   // ฿ per gram
	CurrentSellPerGram  float64            `json:"current_sell_per_gram"`  // ฿ per gram (for forward planning)
	CurrentValue        float64            `json:"current_value"`          // ฿
	CostBasisValue      float64            `json:"cost_basis_value"`       // ฿
	UnrealizedPnL       float64            `json:"unrealized_pnl"`         // ฿
	UnrealizedPnLPercent float64           `json:"unrealized_pnl_percent"` // %
}

// GetStatement assembles the customer-facing statement.
func (s *Service) GetStatement(ctx context.Context, accountID primitive.ObjectID) (*Statement, error) {
	account, err := s.goldSavingRepo.GetByID(ctx, accountID)
	if err != nil || account == nil {
		return nil, errors.New("account not found")
	}
	gp, _ := s.goldPriceRepo.GetCurrent(ctx)
	var buyPerBaht, buyPerGram, sellPerGram float64
	if gp != nil {
		buyPerBaht = gp.GoldOrnamentBuy
		buyPerGram = gp.GoldOrnamentBuy / BahtPerGramOrnament
		sellPerGram = gp.GoldOrnamentSell / BahtPerGramOrnament
	}
	currentValue := utils.RoundBaht(account.GoldWeight * buyPerGram)
	costBasis := account.CostBasis()
	pnl := utils.RoundBaht(currentValue - costBasis)
	pnlPct := 0.0
	if costBasis > 0 {
		pnlPct = (pnl / costBasis) * 100
	}
	return &Statement{
		Account:              account,
		GoldWeight:           account.GoldWeight,
		CurrentBuyPrice:      buyPerBaht,
		CurrentBuyPerGram:    utils.RoundBaht(buyPerGram),
		CurrentSellPerGram:   utils.RoundBaht(sellPerGram),
		CurrentValue:         currentValue,
		CostBasisValue:       costBasis,
		UnrealizedPnL:        pnl,
		UnrealizedPnLPercent: pnlPct,
	}, nil
}

// loadActiveAccountWithPrices is the shared prelude for deposit/withdraw paths:
// fetch account, ensure it is active, and pull current per-gram sell+buy prices.
func (s *Service) loadActiveAccountWithPrices(ctx context.Context, accountID primitive.ObjectID) (*entity.GoldSaving, float64, float64, error) {
	account, err := s.goldSavingRepo.GetByID(ctx, accountID)
	if err != nil || account == nil {
		return nil, 0, 0, errors.New("account not found")
	}
	if account.Status != entity.GoldSavingStatusActive {
		return nil, 0, 0, errors.New("account is not active")
	}
	gp, err := s.goldPriceRepo.GetCurrent(ctx)
	if err != nil || gp == nil {
		return nil, 0, 0, errors.New("failed to get current gold price")
	}
	if gp.GoldOrnamentSell <= 0 || gp.GoldOrnamentBuy <= 0 {
		return nil, 0, 0, errors.New("invalid gold price configured")
	}
	sellPerGram := gp.GoldOrnamentSell / BahtPerGramOrnament
	buyPerGram := gp.GoldOrnamentBuy / BahtPerGramOrnament
	return account, sellPerGram, buyPerGram, nil
}
