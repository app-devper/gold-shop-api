package sale

import (
	"context"
	"testing"

	"github.com/devper-gold/gold-shop-api/internal/domain/entity"
	"github.com/devper-gold/gold-shop-api/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCreateSale(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		branchID := primitive.NewObjectID()
		userID := primitive.NewObjectID()
		productID := primitive.NewObjectID()

		mockSaleRepo := new(testutils.MockSaleRepository)
		mockProductRepo := new(testutils.MockProductRepository)
		mockCustomerRepo := new(testutils.MockCustomerRepository)
		mockBranchRepo := new(testutils.MockBranchRepository)
		mockUserRepo := new(testutils.MockUserRepository)
		mockItemRepo := new(testutils.MockProductItemRepository)
		mockPriceRepo := new(testutils.MockGoldPriceRepository)
		mockStockLogRepo := new(testutils.MockStockLogRepository)

		service := NewService(mockSaleRepo, mockProductRepo, mockItemRepo, mockPriceRepo, mockStockLogRepo, mockCustomerRepo, mockBranchRepo, mockUserRepo)

		mockPriceRepo.On("GetCurrent", ctx).Return(&entity.GoldPrice{GoldBarBuy: 30000, GoldBarSell: 31000}, nil)

		branch := &entity.Branch{ID: branchID, Code: "B001", Name: "Test Branch"}
		product := &entity.Product{
			ID:        productID,
			BranchID:  branchID,
			SKU:       "GOLD-965-1G",
			Name:      "Gold 96.5% 1g",
			StockType: entity.StockTypeWeight,
			Weight:    10.0,
			Price:     3000,
			Cost:      2500,
			Status:    entity.ProductStatusAvailable,
		}

		mockBranchRepo.On("GetByID", ctx, branchID).Return(branch, nil)
		mockSaleRepo.On("GenerateSaleNumber", ctx, "B001").Return("S20240001", nil)
		mockProductRepo.On("GetByID", ctx, productID).Return(product, nil)
		mockProductRepo.On("Update", ctx, mock.AnythingOfType("*entity.Product")).Return(nil)
		mockStockLogRepo.On("Create", ctx, mock.AnythingOfType("*entity.StockLog")).Return(nil)
		mockSaleRepo.On("Create", ctx, mock.AnythingOfType("*entity.Sale")).Return(nil)

		input := CreateSaleInput{
			BranchID: branchID,
			UserID:   userID,
			SaleType: entity.SaleTypeSell,
			Items: []SaleItemInput{
				{
					ProductID:  productID.Hex(),
					PriceLevel: "A",
					Weight:     1.0,
				},
			},
			Payments: []PaymentInput{
				{
					Method: entity.PaymentMethodCash,
					Amount: 3000,
				},
			},
		}

		sale, err := service.Create(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, sale)
		assert.Equal(t, "S20240001", sale.SaleNumber)
		assert.Equal(t, 2033.5869850432957, sale.NetTotal)
		assert.Equal(t, 2500.0, sale.Items[0].Cost)

		mockBranchRepo.AssertExpectations(t)
	})

	t.Run("BranchNotFound", func(t *testing.T) {
		branchID := primitive.NewObjectID()
		mockBranchRepo := new(testutils.MockBranchRepository)
		service := NewService(nil, nil, nil, nil, nil, nil, mockBranchRepo, nil)

		mockBranchRepo.On("GetByID", ctx, branchID).Return(nil, nil).Once()

		input := CreateSaleInput{BranchID: branchID}
		sale, err := service.Create(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, sale)
		assert.Equal(t, "branch not found", err.Error())
	})
}

func TestGenerateReceipt(t *testing.T) {
	ctx := context.Background()
	branchID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	saleID := primitive.NewObjectID()

	mockSaleRepo := new(testutils.MockSaleRepository)
	mockProductRepo := new(testutils.MockProductRepository)
	mockCustomerRepo := new(testutils.MockCustomerRepository)
	mockBranchRepo := new(testutils.MockBranchRepository)
	mockUserRepo := new(testutils.MockUserRepository)

	service := NewService(mockSaleRepo, mockProductRepo, nil, nil, nil, mockCustomerRepo, mockBranchRepo, mockUserRepo)

	t.Run("Success", func(t *testing.T) {
		sale := &entity.Sale{
			ID:         saleID,
			BranchID:   branchID,
			UserID:     userID,
			SaleNumber: "S20240001",
		}
		branch := &entity.Branch{ID: branchID, Name: "Main Branch"}
		user := &entity.User{ID: userID, FullName: "John Doe"}

		mockSaleRepo.On("GetByID", ctx, saleID).Return(sale, nil)
		mockBranchRepo.On("GetByID", ctx, branchID).Return(branch, nil)
		mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

		receipt, err := service.GenerateReceipt(ctx, saleID)

		assert.NoError(t, err)
		assert.NotNil(t, receipt)
		assert.Equal(t, "S20240001", receipt.Sale.SaleNumber)
		assert.Equal(t, "Main Branch", receipt.BranchName)
		assert.Equal(t, "John Doe", receipt.CashierName)

		mockUserRepo.AssertExpectations(t)
	})
}

func floatPtr(f float64) *float64 {
	return &f
}
