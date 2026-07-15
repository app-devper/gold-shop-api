package gold_saving

import (
	"context"
	"testing"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// goldPriceForTests — bar 30,000 ฿/baht (sell) and 29,000 ฿/baht (buy).
// per-gram price = baht-price / 15.244 → sell ≈ 1968.04, buy ≈ 1902.39
func goldPriceForTests() *entity.GoldPrice {
	return &entity.GoldPrice{
		GoldBarSell: 30000,
		GoldBarBuy:  29000,
	}
}

func TestOpen(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		branchID := primitive.NewObjectID()
		customerID := primitive.NewObjectID()

		savingRepo := new(testutils.MockGoldSavingRepository)
		branchRepo := new(testutils.MockBranchRepository)
		customerRepo := new(testutils.MockCustomerRepository)
		s := NewService(savingRepo, nil, branchRepo, customerRepo)

		branch := &entity.Branch{ID: branchID, Code: "B001"}
		customer := &entity.Customer{ID: customerID, FullName: "Cust"}

		branchRepo.On("GetByID", ctx, branchID).Return(branch, nil)
		customerRepo.On("GetByID", ctx, customerID).Return(customer, nil)
		savingRepo.On("GenerateAccountNumber", ctx, "B001").Return("GS-B001-0001", nil)
		savingRepo.On("Create", ctx, mock.AnythingOfType("*entity.GoldSaving")).Return(nil)

		account, err := s.Open(ctx, OpenInput{BranchID: branchID, CustomerID: customerID})

		assert.NoError(t, err)
		assert.Equal(t, "GS-B001-0001", account.AccountNumber)
		assert.Equal(t, entity.GoldSavingStatusActive, account.Status)
		assert.Equal(t, 0.0, account.GoldWeight)
	})

	t.Run("CustomerNotFound", func(t *testing.T) {
		branchRepo := new(testutils.MockBranchRepository)
		customerRepo := new(testutils.MockCustomerRepository)
		s := NewService(nil, nil, branchRepo, customerRepo)
		branchID := primitive.NewObjectID()
		customerID := primitive.NewObjectID()
		branchRepo.On("GetByID", ctx, branchID).Return(&entity.Branch{ID: branchID, Code: "B001"}, nil)
		customerRepo.On("GetByID", ctx, customerID).Return(nil, nil)
		_, err := s.Open(ctx, OpenInput{BranchID: branchID, CustomerID: customerID})
		assert.EqualError(t, err, "customer not found")
	})
}

func TestDepositCash(t *testing.T) {
	ctx := context.Background()
	accountID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	t.Run("Success_DerivesWeightAtSellPrice", func(t *testing.T) {
		savingRepo := new(testutils.MockGoldSavingRepository)
		priceRepo := new(testutils.MockGoldPriceRepository)
		s := NewService(savingRepo, priceRepo, nil, nil)

		account := &entity.GoldSaving{
			ID:     accountID,
			Status: entity.GoldSavingStatusActive,
		}
		savingRepo.On("GetByID", ctx, accountID).Return(account, nil)
		priceRepo.On("GetCurrent", ctx).Return(goldPriceForTests(), nil)
		savingRepo.On("Update", ctx, mock.AnythingOfType("*entity.GoldSaving")).Return(nil)

		// Bar sell price = 30000 / 15.244 = 1968.04 ฿/g; ฿2000 → ~1.0163 g
		got, err := s.DepositCash(ctx, accountID, 2000, userID)
		assert.NoError(t, err)
		assert.InDelta(t, 1.0163, got.GoldWeight, 0.001)
		assert.Equal(t, 2000.0, got.TotalDepositValue)
		assert.Len(t, got.Transactions, 1)
		tx := got.Transactions[0]
		assert.Equal(t, entity.TxDeposit, tx.Type)
		assert.Equal(t, entity.TxModeCash, tx.Mode)
		assert.True(t, tx.GoldWeightDelta > 0)
	})

	t.Run("BelowMinimum", func(t *testing.T) {
		savingRepo := new(testutils.MockGoldSavingRepository)
		priceRepo := new(testutils.MockGoldPriceRepository)
		s := NewService(savingRepo, priceRepo, nil, nil)

		account := &entity.GoldSaving{
			Status:         entity.GoldSavingStatusActive,
			MinDepositCash: 5000,
		}
		savingRepo.On("GetByID", ctx, accountID).Return(account, nil)
		priceRepo.On("GetCurrent", ctx).Return(goldPriceForTests(), nil)

		_, err := s.DepositCash(ctx, accountID, 1000, userID)
		assert.ErrorContains(t, err, "minimum cash deposit")
	})
}

func TestDepositGold(t *testing.T) {
	ctx := context.Background()
	accountID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	savingRepo := new(testutils.MockGoldSavingRepository)
	priceRepo := new(testutils.MockGoldPriceRepository)
	s := NewService(savingRepo, priceRepo, nil, nil)

	account := &entity.GoldSaving{Status: entity.GoldSavingStatusActive}
	savingRepo.On("GetByID", ctx, accountID).Return(account, nil)
	priceRepo.On("GetCurrent", ctx).Return(goldPriceForTests(), nil)
	savingRepo.On("Update", ctx, mock.AnythingOfType("*entity.GoldSaving")).Return(nil)

	got, err := s.DepositGold(ctx, accountID, 2.5, userID)
	assert.NoError(t, err)
	assert.InDelta(t, 2.5, got.GoldWeight, 1e-6)
	// 2.5g × (30000/15.244) ≈ 4919.97 ฿ cost basis
	assert.InDelta(t, 4919.97, got.TotalDepositValue, 0.5)
}

func TestWithdrawCash_UsesBuyPriceAndChecksBalance(t *testing.T) {
	ctx := context.Background()
	accountID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	savingRepo := new(testutils.MockGoldSavingRepository)
	priceRepo := new(testutils.MockGoldPriceRepository)
	s := NewService(savingRepo, priceRepo, nil, nil)

	// Pre-existing balance: 1g
	account := &entity.GoldSaving{Status: entity.GoldSavingStatusActive, GoldWeight: 1.0}
	savingRepo.On("GetByID", ctx, accountID).Return(account, nil)
	priceRepo.On("GetCurrent", ctx).Return(goldPriceForTests(), nil)
	savingRepo.On("Update", ctx, mock.AnythingOfType("*entity.GoldSaving")).Return(nil)

	// Bar buy price = 29000 / 15.244 = 1902.39 ฿/g.
	// Withdraw ฿1500 → weight 0.7885g → remaining ~0.2115g
	got, err := s.WithdrawCash(ctx, accountID, 1500, userID)
	assert.NoError(t, err)
	assert.InDelta(t, 0.2115, got.GoldWeight, 0.001)
	assert.Equal(t, 1500.0, got.TotalWithdrawValue)
}

func TestWithdrawCash_InsufficientBalance(t *testing.T) {
	ctx := context.Background()
	accountID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	savingRepo := new(testutils.MockGoldSavingRepository)
	priceRepo := new(testutils.MockGoldPriceRepository)
	s := NewService(savingRepo, priceRepo, nil, nil)

	account := &entity.GoldSaving{Status: entity.GoldSavingStatusActive, GoldWeight: 0.1}
	savingRepo.On("GetByID", ctx, accountID).Return(account, nil)
	priceRepo.On("GetCurrent", ctx).Return(goldPriceForTests(), nil)

	// Want ฿10,000 ≈ 5.226g but only 0.1g
	_, err := s.WithdrawCash(ctx, accountID, 10000, userID)
	assert.ErrorIs(t, err, entity.ErrInsufficientBalance)
}

func TestAdjust_RequiresNoteAndRecordsAudit(t *testing.T) {
	ctx := context.Background()
	accountID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	savingRepo := new(testutils.MockGoldSavingRepository)
	priceRepo := new(testutils.MockGoldPriceRepository)
	s := NewService(savingRepo, priceRepo, nil, nil)

	account := &entity.GoldSaving{Status: entity.GoldSavingStatusActive, GoldWeight: 1.0}
	savingRepo.On("GetByID", ctx, accountID).Return(account, nil)
	priceRepo.On("GetCurrent", ctx).Return(goldPriceForTests(), nil)
	savingRepo.On("Update", ctx, mock.AnythingOfType("*entity.GoldSaving")).Return(nil)

	// Missing note → reject
	_, err := s.Adjust(ctx, AdjustInput{AccountID: accountID, WeightDelta: -0.1, By: userID})
	assert.EqualError(t, err, "note is required for adjustment")

	got, err := s.Adjust(ctx, AdjustInput{
		AccountID: accountID, WeightDelta: -0.1, Note: "physical loss reconciliation", By: userID,
	})
	assert.NoError(t, err)
	assert.InDelta(t, 0.9, got.GoldWeight, 1e-6)
	assert.Len(t, got.Transactions, 1)
	tx := got.Transactions[0]
	assert.Equal(t, entity.TxAdjust, tx.Type)
	assert.Equal(t, "physical loss reconciliation", tx.Note)
}

func TestStatement_MarkToMarketAndPnL(t *testing.T) {
	ctx := context.Background()
	accountID := primitive.NewObjectID()

	savingRepo := new(testutils.MockGoldSavingRepository)
	priceRepo := new(testutils.MockGoldPriceRepository)
	s := NewService(savingRepo, priceRepo, nil, nil)

	// Customer deposited ฿5000 cash earlier; balance now ~2.526g.
	// At bar buy price 1902.39, current value ≈ 4805.43 → PnL ≈ -194.57
	account := &entity.GoldSaving{
		ID:                 accountID,
		GoldWeight:         2.526,
		TotalDepositValue:  5000,
		TotalDepositWeight: 2.526,
	}
	savingRepo.On("GetByID", ctx, accountID).Return(account, nil)
	priceRepo.On("GetCurrent", ctx).Return(goldPriceForTests(), nil)

	st, err := s.GetStatement(ctx, accountID)
	assert.NoError(t, err)
	assert.InDelta(t, 4805.43, st.CurrentValue, 1)
	assert.InDelta(t, -194.57, st.UnrealizedPnL, 1)
	assert.True(t, st.UnrealizedPnLPercent < 0)
}

func TestClose_RequiresZeroBalance(t *testing.T) {
	ctx := context.Background()
	accountID := primitive.NewObjectID()

	savingRepo := new(testutils.MockGoldSavingRepository)
	s := NewService(savingRepo, nil, nil, nil)

	// non-zero balance → reject
	savingRepo.On("GetByID", ctx, accountID).Return(&entity.GoldSaving{
		Status: entity.GoldSavingStatusActive, GoldWeight: 0.5,
	}, nil).Once()
	_, err := s.Close(ctx, accountID)
	assert.ErrorContains(t, err, "still has balance")

	// zero balance → ok
	savingRepo.On("GetByID", ctx, accountID).Return(&entity.GoldSaving{
		Status: entity.GoldSavingStatusActive, GoldWeight: 0,
	}, nil).Once()
	savingRepo.On("Update", ctx, mock.AnythingOfType("*entity.GoldSaving")).Return(nil).Once()
	got, err := s.Close(ctx, accountID)
	assert.NoError(t, err)
	assert.Equal(t, entity.GoldSavingStatusClosed, got.Status)
}

func TestGetByBranchIDAttachesCustomerNames(t *testing.T) {
	ctx := context.Background()
	branchID := primitive.NewObjectID()
	customerID := primitive.NewObjectID()

	savingRepo := new(testutils.MockGoldSavingRepository)
	customerRepo := new(testutils.MockCustomerRepository)
	s := NewService(savingRepo, nil, nil, customerRepo)

	accounts := []*entity.GoldSaving{{ID: primitive.NewObjectID(), CustomerID: customerID}}
	savingRepo.On("GetByBranchID", ctx, branchID, []entity.GoldSavingStatus(nil)).Return(accounts, nil)
	customerRepo.On("GetNamesByIDs", ctx, []primitive.ObjectID{customerID}).Return(map[primitive.ObjectID]string{
		customerID: "สมชาย ใจดี",
	}, nil)

	result, err := s.GetByBranchID(ctx, branchID, nil)

	assert.NoError(t, err)
	assert.Equal(t, "สมชาย ใจดี", result[0].CustomerName)
	customerRepo.AssertExpectations(t)
}

func TestGetByIDAttachesCustomerName(t *testing.T) {
	ctx := context.Background()
	accountID := primitive.NewObjectID()
	customerID := primitive.NewObjectID()

	savingRepo := new(testutils.MockGoldSavingRepository)
	customerRepo := new(testutils.MockCustomerRepository)
	s := NewService(savingRepo, nil, nil, customerRepo)

	savingRepo.On("GetByID", ctx, accountID).Return(&entity.GoldSaving{ID: accountID, CustomerID: customerID}, nil)
	customerRepo.On("GetNamesByIDs", ctx, []primitive.ObjectID{customerID}).Return(map[primitive.ObjectID]string{
		customerID: "สมปอง มั่งมี",
	}, nil)

	result, err := s.GetByID(ctx, accountID)

	assert.NoError(t, err)
	assert.Equal(t, "สมปอง มั่งมี", result.CustomerName)
}

func TestSearchMatchesAccountNumberAndCustomerName(t *testing.T) {
	ctx := context.Background()
	branchID := primitive.NewObjectID()
	customerID := primitive.NewObjectID()

	savingRepo := new(testutils.MockGoldSavingRepository)
	customerRepo := new(testutils.MockCustomerRepository)
	s := NewService(savingRepo, nil, nil, customerRepo)

	customerRepo.On("Search", ctx, "สมชาย", 5).Return([]*entity.Customer{
		{ID: customerID, FullName: "สมชาย ใจดี"},
	}, nil)
	accounts := []*entity.GoldSaving{{ID: primitive.NewObjectID(), CustomerID: customerID}}
	savingRepo.On("Search", ctx, branchID, "สมชาย", []primitive.ObjectID{customerID}, 5).Return(accounts, nil)
	customerRepo.On("GetNamesByIDs", ctx, []primitive.ObjectID{customerID}).Return(map[primitive.ObjectID]string{
		customerID: "สมชาย ใจดี",
	}, nil)

	result, err := s.Search(ctx, branchID, "สมชาย", 5)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "สมชาย ใจดี", result[0].CustomerName)
	savingRepo.AssertExpectations(t)
	customerRepo.AssertExpectations(t)
}
