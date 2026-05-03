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

func TestOpenAccount(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		branchID := primitive.NewObjectID()
		customerID := primitive.NewObjectID()

		mockSavingRepo := new(testutils.MockGoldSavingRepository)
		mockBranchRepo := new(testutils.MockBranchRepository)
		mockCustomerRepo := new(testutils.MockCustomerRepository)
		service := NewService(mockSavingRepo, nil, mockBranchRepo, mockCustomerRepo)

		branch := &entity.Branch{ID: branchID, Code: "B001", Name: "Test Branch"}
		customer := &entity.Customer{ID: customerID, FullName: "Test Customer"}

		mockBranchRepo.On("GetByID", ctx, branchID).Return(branch, nil)
		mockCustomerRepo.On("GetByID", ctx, customerID).Return(customer, nil)
		mockSavingRepo.On("GenerateAccountNumber", ctx, "B001").Return("GS20240001", nil)
		mockSavingRepo.On("Create", ctx, mock.AnythingOfType("*entity.GoldSaving")).Return(nil)

		account, err := service.OpenAccount(ctx, branchID, customerID, entity.GoldSavingByWeight, 100, 50)

		assert.NoError(t, err)
		assert.NotNil(t, account)
		assert.Equal(t, "GS20240001", account.AccountNumber)
		assert.Equal(t, entity.GoldSavingByWeight, account.SavingType)

		mockBranchRepo.AssertExpectations(t)
		mockCustomerRepo.AssertExpectations(t)
		mockSavingRepo.AssertExpectations(t)
	})

	t.Run("BranchNotFound", func(t *testing.T) {
		branchID := primitive.NewObjectID()
		mockBranchRepo := new(testutils.MockBranchRepository)
		service := NewService(nil, nil, mockBranchRepo, nil)

		mockBranchRepo.On("GetByID", ctx, branchID).Return(nil, nil).Once()

		account, err := service.OpenAccount(ctx, branchID, primitive.NewObjectID(), entity.GoldSavingByWeight, 100, 50)

		assert.Error(t, err)
		assert.Nil(t, account)
		assert.Equal(t, "branch not found", err.Error())
	})

	t.Run("CustomerNotFound", func(t *testing.T) {
		branchID := primitive.NewObjectID()
		customerID := primitive.NewObjectID()
		mockBranchRepo := new(testutils.MockBranchRepository)
		mockCustomerRepo := new(testutils.MockCustomerRepository)
		service := NewService(nil, nil, mockBranchRepo, mockCustomerRepo)

		branch := &entity.Branch{ID: branchID, Code: "B001"}
		mockBranchRepo.On("GetByID", ctx, branchID).Return(branch, nil)
		mockCustomerRepo.On("GetByID", ctx, customerID).Return(nil, nil)

		account, err := service.OpenAccount(ctx, branchID, customerID, entity.GoldSavingByWeight, 0, 0)

		assert.Error(t, err)
		assert.Nil(t, account)
		assert.Equal(t, "customer not found", err.Error())
	})

	t.Run("InvalidSavingType", func(t *testing.T) {
		service := NewService(nil, nil, nil, nil)
		_, err := service.OpenAccount(ctx, primitive.NewObjectID(), primitive.NewObjectID(), "bogus", 0, 0)
		assert.EqualError(t, err, "invalid saving type")
	})
}

func TestDeposit(t *testing.T) {
	ctx := context.Background()
	accountID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	t.Run("Success", func(t *testing.T) {
		mockSavingRepo := new(testutils.MockGoldSavingRepository)
		mockPriceRepo := new(testutils.MockGoldPriceRepository)
		service := NewService(mockSavingRepo, mockPriceRepo, nil, nil)

		account := &entity.GoldSaving{
			ID:         accountID,
			SavingType: entity.GoldSavingByMoney,
			Status:     entity.GoldSavingStatusActive,
			MinDeposit: 100,
		}
		goldPrice := &entity.GoldPrice{
			GoldOrnamentSell: 30000,
		}

		mockSavingRepo.On("GetByID", ctx, accountID).Return(account, nil)
		mockPriceRepo.On("GetCurrent", ctx).Return(goldPrice, nil)
		mockSavingRepo.On("Update", ctx, mock.AnythingOfType("*entity.GoldSaving")).Return(nil)

		updatedAccount, err := service.Deposit(ctx, accountID, 3000, userID)

		assert.NoError(t, err)
		assert.NotNil(t, updatedAccount)
		assert.InDelta(t, 0.1, updatedAccount.GoldBalance, 0.0001) // 3000 / 30000 = 0.1

		mockSavingRepo.AssertExpectations(t)
		mockPriceRepo.AssertExpectations(t)
	})

	t.Run("AccountNotFound", func(t *testing.T) {
		mockSavingRepo := new(testutils.MockGoldSavingRepository)
		service := NewService(mockSavingRepo, nil, nil, nil)

		mockSavingRepo.On("GetByID", ctx, accountID).Return(nil, nil).Once()

		updatedAccount, err := service.Deposit(ctx, accountID, 3000, userID)

		assert.Error(t, err)
		assert.Nil(t, updatedAccount)
		assert.Equal(t, "account not found", err.Error())
	})

	t.Run("AmountZeroRejected", func(t *testing.T) {
		service := NewService(nil, nil, nil, nil)
		_, err := service.Deposit(ctx, accountID, 0, userID)
		assert.EqualError(t, err, "amount must be greater than zero")
	})

	t.Run("BelowMinimumWithUnit", func(t *testing.T) {
		mockSavingRepo := new(testutils.MockGoldSavingRepository)
		service := NewService(mockSavingRepo, nil, nil, nil)

		mockSavingRepo.On("GetByID", ctx, accountID).Return(&entity.GoldSaving{
			Status:     entity.GoldSavingStatusActive,
			SavingType: entity.GoldSavingByMoney,
			MinDeposit: 1000,
		}, nil).Once()

		_, err := service.Deposit(ctx, accountID, 50, userID)
		assert.ErrorContains(t, err, "amount is below minimum deposit")
		assert.ErrorContains(t, err, "฿")
	})
}
