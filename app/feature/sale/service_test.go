package sale

import (
	"context"
	"testing"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCreateSale(t *testing.T) {
	ctx := context.Background()

	t.Run("Success_OrnamentPiece", func(t *testing.T) {
		branchID := primitive.NewObjectID()
		userID := primitive.NewObjectID()
		productID := primitive.NewObjectID()
		itemID := primitive.NewObjectID()

		mockSaleRepo := new(testutils.MockSaleRepository)
		mockProductRepo := new(testutils.MockProductRepository)
		mockCustomerRepo := new(testutils.MockCustomerRepository)
		mockBranchRepo := new(testutils.MockBranchRepository)
		mockUserRepo := new(testutils.MockUserRepository)
		mockItemRepo := new(testutils.MockProductItemRepository)
		mockPriceRepo := new(testutils.MockGoldPriceRepository)
		mockStockLogRepo := new(testutils.MockStockLogRepository)
		mockTxManager := new(testutils.MockTransactionManager)
		service := NewService(mockSaleRepo, mockProductRepo, mockItemRepo, mockPriceRepo, mockStockLogRepo, mockCustomerRepo, mockBranchRepo, mockUserRepo, mockTxManager)

		mockPriceRepo.On("GetCurrent", ctx).Return(&entity.GoldPrice{
			GoldBarBuy: 30000, GoldBarSell: 31000,
			GoldOrnamentBuy: 29000, GoldOrnamentSell: 30500,
			Source: "api",
		}, nil)

		branch := &entity.Branch{ID: branchID, Code: "B001"}
		product := &entity.Product{
			ID: productID, BranchID: branchID, SKU: "RING-001",
			Kind: entity.KindOrnament, GoldType: "96.5%", Name: "แหวนทองคำ",
			IsActive: true,
		}
		item := &entity.ProductItem{
			ID: itemID, ProductID: productID, BranchID: branchID,
			Barcode: "BC-001", WeightGrams: 7.6, LaborCost: 500, Cost: 14000,
			Status: entity.ProductStatusAvailable,
		}

		mockBranchRepo.On("GetByID", ctx, branchID).Return(branch, nil)
		mockSaleRepo.On("GenerateSaleNumber", ctx, "B001", entity.SaleTypeSell).Return("S-001", nil)
		mockProductRepo.On("GetByID", ctx, productID).Return(product, nil)
		mockItemRepo.On("GetByID", ctx, itemID).Return(item, nil)
		mockItemRepo.On("Update", ctx, mock.AnythingOfType("*entity.ProductItem")).Return(nil)
		mockStockLogRepo.On("Create", ctx, mock.AnythingOfType("*entity.StockLog")).Return(nil)
		mockSaleRepo.On("Create", ctx, mock.AnythingOfType("*entity.Sale")).Return(nil)

		input := CreateSaleInput{
			BranchID: branchID,
			UserID:   userID,
			SaleType: entity.SaleTypeSell,
			Items: []SaleItemInput{{
				ProductID:     productID.Hex(),
				ProductItemID: itemID.Hex(),
				PriceLevel:    "A",
			}},
			Payments: []PaymentInput{{Method: entity.PaymentMethodCash, Amount: 16000}},
		}

		sale, err := service.Create(ctx, input)
		assert.NoError(t, err)
		assert.NotNil(t, sale)
		// ornament uses sell ornament price = 30500/15.16 ≈ 2012.53 ฿/g; 7.6g + labor 500
		// total ≈ 7.6 × 2012.53 + 500 = 15795.24
		assert.InDelta(t, 15795.24, sale.NetTotal, 5)
		assert.Equal(t, 14000.0, sale.Items[0].Cost)
		assert.Equal(t, "api", sale.GoldPrice.Source)
		assert.Equal(t, "BC-001", sale.Items[0].Barcode)
		assert.InDelta(t, 2011.87, sale.Items[0].PricePerGram, 0.01)

		mockItemRepo.AssertExpectations(t)
	})

	t.Run("AutoCalculatesOldGoldBuybackWithDeduction", func(t *testing.T) {
		branchID := primitive.NewObjectID()
		userID := primitive.NewObjectID()

		mockSaleRepo := new(testutils.MockSaleRepository)
		mockProductRepo := new(testutils.MockProductRepository)
		mockCustomerRepo := new(testutils.MockCustomerRepository)
		mockBranchRepo := new(testutils.MockBranchRepository)
		mockUserRepo := new(testutils.MockUserRepository)
		mockItemRepo := new(testutils.MockProductItemRepository)
		mockPriceRepo := new(testutils.MockGoldPriceRepository)
		mockStockLogRepo := new(testutils.MockStockLogRepository)
		mockTxManager := new(testutils.MockTransactionManager)
		service := NewService(mockSaleRepo, mockProductRepo, mockItemRepo, mockPriceRepo, mockStockLogRepo, mockCustomerRepo, mockBranchRepo, mockUserRepo, mockTxManager)

		mockPriceRepo.On("GetCurrent", ctx).Return(&entity.GoldPrice{
			GoldBarBuy: 30000, GoldBarSell: 31000,
			GoldOrnamentBuy: 29000, GoldOrnamentSell: 30500,
			Source: "api",
		}, nil)
		mockBranchRepo.On("GetByID", ctx, branchID).Return(&entity.Branch{ID: branchID, Code: "B001"}, nil)
		mockSaleRepo.On("GenerateSaleNumber", ctx, "B001", entity.SaleTypeBuyOld).Return("S-002", nil)
		mockSaleRepo.On("Create", ctx, mock.AnythingOfType("*entity.Sale")).Return(nil)

		sale, err := service.Create(ctx, CreateSaleInput{
			BranchID: branchID,
			UserID:   userID,
			SaleType: entity.SaleTypeBuyOld,
			OldGoldItems: []OldGoldInput{{
				Description:      "สร้อยเก่า",
				GoldType:         "96.5%",
				Kind:             entity.KindOrnament,
				Weight:           7.6,
				DeductionPercent: 3,
			}},
			Payments: []PaymentInput{{Method: entity.PaymentMethodCash, Amount: 0}},
		})

		assert.NoError(t, err)
		assert.NotNil(t, sale)
		assert.Len(t, sale.OldGoldItems, 1)
		assert.InDelta(t, 14102.11, sale.OldGoldItems[0].Total, 0.01)
		assert.Equal(t, 0.0, sale.NetTotal)
		assert.Equal(t, "api", sale.GoldPrice.Source)
	})

	t.Run("RequireProductItemID", func(t *testing.T) {
		branchID := primitive.NewObjectID()
		productID := primitive.NewObjectID()

		mockProductRepo := new(testutils.MockProductRepository)
		mockBranchRepo := new(testutils.MockBranchRepository)
		mockPriceRepo := new(testutils.MockGoldPriceRepository)
		service := NewService(nil, mockProductRepo, nil, mockPriceRepo, nil, nil, mockBranchRepo, nil, new(testutils.MockTransactionManager))

		mockBranchRepo.On("GetByID", ctx, branchID).Return(&entity.Branch{ID: branchID, Code: "B"}, nil)
		mockPriceRepo.On("GetCurrent", ctx).Return(&entity.GoldPrice{GoldBarBuy: 30000, GoldBarSell: 31000, GoldOrnamentBuy: 29000, GoldOrnamentSell: 30500}, nil)
		mockProductRepo.On("GetByID", ctx, productID).Return(&entity.Product{
			ID: productID, BranchID: branchID, Kind: entity.KindBar, IsActive: true,
		}, nil)

		_, err := service.Create(ctx, CreateSaleInput{
			BranchID: branchID,
			SaleType: entity.SaleTypeSell,
			Items:    []SaleItemInput{{ProductID: productID.Hex()}}, // missing ProductItemID
			Payments: []PaymentInput{{Method: entity.PaymentMethodCash, Amount: 1}},
		})
		assert.ErrorContains(t, err, "product item ID is required")
	})

	t.Run("BranchNotFound", func(t *testing.T) {
		branchID := primitive.NewObjectID()
		mockBranchRepo := new(testutils.MockBranchRepository)
		service := NewService(nil, nil, nil, nil, nil, nil, mockBranchRepo, nil, new(testutils.MockTransactionManager))

		mockBranchRepo.On("GetByID", ctx, branchID).Return(nil, nil).Once()
		_, err := service.Create(ctx, CreateSaleInput{BranchID: branchID})
		assert.EqualError(t, err, "branch not found")
	})
}

func TestGenerateReceipt(t *testing.T) {
	ctx := context.Background()
	branchID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	saleID := primitive.NewObjectID()

	mockSaleRepo := new(testutils.MockSaleRepository)
	mockBranchRepo := new(testutils.MockBranchRepository)
	mockUserRepo := new(testutils.MockUserRepository)
	service := NewService(mockSaleRepo, nil, nil, nil, nil, nil, mockBranchRepo, mockUserRepo, new(testutils.MockTransactionManager))

	sale := &entity.Sale{ID: saleID, BranchID: branchID, UserID: userID, SaleNumber: "S-001"}
	branch := &entity.Branch{ID: branchID, Name: "Main"}
	user := &entity.User{ID: userID, FullName: "John"}

	mockSaleRepo.On("GetByID", ctx, saleID).Return(sale, nil)
	mockBranchRepo.On("GetByID", ctx, branchID).Return(branch, nil)
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	receipt, err := service.GenerateReceipt(ctx, saleID)
	assert.NoError(t, err)
	assert.Equal(t, "S-001", receipt.Sale.SaleNumber)
	assert.Equal(t, "Main", receipt.BranchName)
	assert.Equal(t, "John", receipt.CashierName)
}
