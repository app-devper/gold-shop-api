package gold_saving

import (
	"context"
	"testing"

	"github.com/devper-gold/gold-shop-api/internal/domain/entity"
	"github.com/devper-gold/gold-shop-api/internal/testutils"
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
		service := NewService(mockSavingRepo, nil, mockBranchRepo)

		branch := &entity.Branch{ID: branchID, Code: "B001", Name: "Test Branch"}

		mockBranchRepo.On("GetByID", ctx, branchID).Return(branch, nil)
		mockSavingRepo.On("GenerateAccountNumber", ctx, "B001").Return("GS20240001", nil)
		mockSavingRepo.On("Create", ctx, mock.AnythingOfType("*entity.GoldSaving")).Return(nil)

		account, err := service.OpenAccount(ctx, branchID, customerID, entity.GoldSavingByWeight, 100, 50)

		assert.NoError(t, err)
		assert.NotNil(t, account)
		assert.Equal(t, "GS20240001", account.AccountNumber)
		assert.Equal(t, entity.GoldSavingByWeight, account.SavingType)

		mockBranchRepo.AssertExpectations(t)
		mockSavingRepo.AssertExpectations(t)
	})

	t.Run("BranchNotFound", func(t *testing.T) {
		branchID := primitive.NewObjectID()
		mockBranchRepo := new(testutils.MockBranchRepository)
		service := NewService(nil, nil, mockBranchRepo)

		mockBranchRepo.On("GetByID", ctx, branchID).Return(nil, nil).Once()

		account, err := service.OpenAccount(ctx, branchID, primitive.NewObjectID(), entity.GoldSavingByWeight, 100, 50)

		assert.Error(t, err)
		assert.Nil(t, account)
		assert.Equal(t, "branch not found", err.Error())
	})
}

func TestDeposit(t *testing.T) {
	ctx := context.Background()
	accountID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	t.Run("Success", func(t *testing.T) {
		mockSavingRepo := new(testutils.MockGoldSavingRepository)
		mockPriceRepo := new(testutils.MockGoldPriceRepository)
		service := NewService(mockSavingRepo, mockPriceRepo, nil)

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
		service := NewService(mockSavingRepo, nil, nil)

		mockSavingRepo.On("GetByID", ctx, accountID).Return(nil, nil).Once()

		updatedAccount, err := service.Deposit(ctx, accountID, 3000, userID)

		assert.Error(t, err)
		assert.Nil(t, updatedAccount)
		assert.Equal(t, "account not found", err.Error())
	})
}
