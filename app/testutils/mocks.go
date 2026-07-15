package testutils

import (
	"context"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockTransactionManager executes fn directly (no real transaction in tests)
type MockTransactionManager struct{}

func (m *MockTransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type MockSaleRepository struct {
	mock.Mock
}

func (m *MockSaleRepository) Create(ctx context.Context, sale *entity.Sale) error {
	args := m.Called(ctx, sale)
	return args.Error(0)
}

func (m *MockSaleRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Sale, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Sale), args.Error(1)
}

func (m *MockSaleRepository) GetBySaleNumber(ctx context.Context, saleNumber string) (*entity.Sale, error) {
	args := m.Called(ctx, saleNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Sale), args.Error(1)
}

func (m *MockSaleRepository) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.SaleStatus, limit, offset int) ([]*entity.Sale, error) {
	args := m.Called(ctx, branchID, status, limit, offset)
	return args.Get(0).([]*entity.Sale), args.Error(1)
}

func (m *MockSaleRepository) GetByCustomerID(ctx context.Context, customerID primitive.ObjectID, limit int) ([]*entity.Sale, error) {
	args := m.Called(ctx, customerID, limit)
	return args.Get(0).([]*entity.Sale), args.Error(1)
}

func (m *MockSaleRepository) GetByDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) ([]*entity.Sale, error) {
	args := m.Called(ctx, branchID, from, to)
	return args.Get(0).([]*entity.Sale), args.Error(1)
}

func (m *MockSaleRepository) Update(ctx context.Context, sale *entity.Sale) error {
	args := m.Called(ctx, sale)
	return args.Error(0)
}

func (m *MockSaleRepository) GenerateSaleNumber(ctx context.Context, branchCode string, saleType entity.SaleType) (string, error) {
	args := m.Called(ctx, branchCode, saleType)
	return args.String(0), args.Error(1)
}

func (m *MockSaleRepository) SumByBranchAndDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) (float64, error) {
	args := m.Called(ctx, branchID, from, to)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockSaleRepository) CountByBranchAndDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) (int64, error) {
	args := m.Called(ctx, branchID, from, to)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSaleRepository) SumCostByBranchAndDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) (float64, error) {
	args := m.Called(ctx, branchID, from, to)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockSaleRepository) GetUnpaidByBranchID(ctx context.Context, branchID primitive.ObjectID) ([]*entity.Sale, error) {
	args := m.Called(ctx, branchID)
	return args.Get(0).([]*entity.Sale), args.Error(1)
}

func (m *MockSaleRepository) GetTopSellingProducts(ctx context.Context, branchID primitive.ObjectID, from, to string, limit int) ([]repository.TopProduct, error) {
	args := m.Called(ctx, branchID, from, to, limit)
	return args.Get(0).([]repository.TopProduct), args.Error(1)
}

func (m *MockSaleRepository) GetEmployeePerformance(ctx context.Context, branchID primitive.ObjectID, from, to string) ([]repository.EmployeePerformance, error) {
	args := m.Called(ctx, branchID, from, to)
	return args.Get(0).([]repository.EmployeePerformance), args.Error(1)
}

func (m *MockSaleRepository) GetSalesTrends(ctx context.Context, branchID primitive.ObjectID, from, to string) ([]repository.SalesTrend, error) {
	args := m.Called(ctx, branchID, from, to)
	return args.Get(0).([]repository.SalesTrend), args.Error(1)
}

type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, product *entity.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *MockProductRepository) GetBySKU(ctx context.Context, sku string) (*entity.Product, error) {
	args := m.Called(ctx, sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *MockProductRepository) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, kind entity.ProductKind, search string, limit, offset int) ([]*entity.Product, error) {
	args := m.Called(ctx, branchID, kind, search, limit, offset)
	return args.Get(0).([]*entity.Product), args.Error(1)
}

func (m *MockProductRepository) Update(ctx context.Context, product *entity.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepository) Count(ctx context.Context, branchID primitive.ObjectID) (int64, error) {
	args := m.Called(ctx, branchID)
	return args.Get(0).(int64), args.Error(1)
}

type MockCustomerRepository struct {
	mock.Mock
}

func (m *MockCustomerRepository) Create(ctx context.Context, customer *entity.Customer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *MockCustomerRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Customer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Customer), args.Error(1)
}

func (m *MockCustomerRepository) GetByMemberCode(ctx context.Context, memberCode string) (*entity.Customer, error) {
	args := m.Called(ctx, memberCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Customer), args.Error(1)
}

func (m *MockCustomerRepository) GetByRFID(ctx context.Context, rfidCard string) (*entity.Customer, error) {
	args := m.Called(ctx, rfidCard)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Customer), args.Error(1)
}

func (m *MockCustomerRepository) GetByPhone(ctx context.Context, phone string) (*entity.Customer, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Customer), args.Error(1)
}

func (m *MockCustomerRepository) GetAll(ctx context.Context, limit, offset int) ([]*entity.Customer, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entity.Customer), args.Error(1)
}

func (m *MockCustomerRepository) GetNamesByIDs(ctx context.Context, ids []primitive.ObjectID) (map[primitive.ObjectID]string, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[primitive.ObjectID]string), args.Error(1)
}

func (m *MockCustomerRepository) Search(ctx context.Context, query string, limit int) ([]*entity.Customer, error) {
	args := m.Called(ctx, query, limit)
	return args.Get(0).([]*entity.Customer), args.Error(1)
}

func (m *MockCustomerRepository) Update(ctx context.Context, customer *entity.Customer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *MockCustomerRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCustomerRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

type MockBranchRepository struct {
	mock.Mock
}

func (m *MockBranchRepository) Create(ctx context.Context, branch *entity.Branch) error {
	args := m.Called(ctx, branch)
	return args.Error(0)
}

func (m *MockBranchRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Branch, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Branch), args.Error(1)
}

func (m *MockBranchRepository) GetByCode(ctx context.Context, code string) (*entity.Branch, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Branch), args.Error(1)
}

func (m *MockBranchRepository) GetAll(ctx context.Context) ([]*entity.Branch, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.Branch), args.Error(1)
}

func (m *MockBranchRepository) Update(ctx context.Context, branch *entity.Branch) error {
	args := m.Called(ctx, branch)
	return args.Error(0)
}

func (m *MockBranchRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByBranchID(ctx context.Context, branchID primitive.ObjectID) ([]*entity.User, error) {
	args := m.Called(ctx, branchID)
	return args.Get(0).([]*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetAll(ctx context.Context) ([]*entity.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockPawnRepository struct {
	mock.Mock
}

func (m *MockPawnRepository) Create(ctx context.Context, pawn *entity.Pawn) error {
	args := m.Called(ctx, pawn)
	return args.Error(0)
}

func (m *MockPawnRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Pawn, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Pawn), args.Error(1)
}

func (m *MockPawnRepository) GetByPawnNumber(ctx context.Context, pawnNumber string) (*entity.Pawn, error) {
	args := m.Called(ctx, pawnNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Pawn), args.Error(1)
}

func (m *MockPawnRepository) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.PawnStatus, limit, offset int) ([]*entity.Pawn, error) {
	args := m.Called(ctx, branchID, status, limit, offset)
	return args.Get(0).([]*entity.Pawn), args.Error(1)
}

func (m *MockPawnRepository) Search(ctx context.Context, branchID primitive.ObjectID, query string, customerIDs []primitive.ObjectID, limit int) ([]*entity.Pawn, error) {
	args := m.Called(ctx, branchID, query, customerIDs, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Pawn), args.Error(1)
}

func (m *MockPawnRepository) GetByCustomerID(ctx context.Context, customerID primitive.ObjectID) ([]*entity.Pawn, error) {
	args := m.Called(ctx, customerID)
	return args.Get(0).([]*entity.Pawn), args.Error(1)
}

func (m *MockPawnRepository) GetDueSoon(ctx context.Context, branchID primitive.ObjectID, days int) ([]*entity.Pawn, error) {
	args := m.Called(ctx, branchID, days)
	return args.Get(0).([]*entity.Pawn), args.Error(1)
}

func (m *MockPawnRepository) GetOverdue(ctx context.Context, branchID primitive.ObjectID) ([]*entity.Pawn, error) {
	args := m.Called(ctx, branchID)
	return args.Get(0).([]*entity.Pawn), args.Error(1)
}

func (m *MockPawnRepository) Update(ctx context.Context, pawn *entity.Pawn) error {
	args := m.Called(ctx, pawn)
	return args.Error(0)
}

func (m *MockPawnRepository) GeneratePawnNumber(ctx context.Context, branchCode string) (string, error) {
	args := m.Called(ctx, branchCode)
	return args.String(0), args.Error(1)
}

func (m *MockPawnRepository) CountByStatus(ctx context.Context, branchID primitive.ObjectID, status entity.PawnStatus) (int64, error) {
	args := m.Called(ctx, branchID, status)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPawnRepository) SumInterestByBranchAndDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) (float64, error) {
	args := m.Called(ctx, branchID, from, to)
	return args.Get(0).(float64), args.Error(1)
}

type MockGoldSavingRepository struct {
	mock.Mock
}

func (m *MockGoldSavingRepository) Create(ctx context.Context, goldSaving *entity.GoldSaving) error {
	args := m.Called(ctx, goldSaving)
	return args.Error(0)
}

func (m *MockGoldSavingRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.GoldSaving, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.GoldSaving), args.Error(1)
}

func (m *MockGoldSavingRepository) GetByAccountNumber(ctx context.Context, accountNumber string) (*entity.GoldSaving, error) {
	args := m.Called(ctx, accountNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.GoldSaving), args.Error(1)
}

func (m *MockGoldSavingRepository) GetByCustomerID(ctx context.Context, customerID primitive.ObjectID) ([]*entity.GoldSaving, error) {
	args := m.Called(ctx, customerID)
	return args.Get(0).([]*entity.GoldSaving), args.Error(1)
}

func (m *MockGoldSavingRepository) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.GoldSavingStatus) ([]*entity.GoldSaving, error) {
	args := m.Called(ctx, branchID, status)
	return args.Get(0).([]*entity.GoldSaving), args.Error(1)
}

func (m *MockGoldSavingRepository) Search(ctx context.Context, branchID primitive.ObjectID, query string, customerIDs []primitive.ObjectID, limit int) ([]*entity.GoldSaving, error) {
	args := m.Called(ctx, branchID, query, customerIDs, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.GoldSaving), args.Error(1)
}

func (m *MockGoldSavingRepository) Update(ctx context.Context, goldSaving *entity.GoldSaving) error {
	args := m.Called(ctx, goldSaving)
	return args.Error(0)
}

func (m *MockGoldSavingRepository) GenerateAccountNumber(ctx context.Context, branchCode string) (string, error) {
	args := m.Called(ctx, branchCode)
	return args.String(0), args.Error(1)
}

func (m *MockGoldSavingRepository) SumBalanceByBranch(ctx context.Context, branchID primitive.ObjectID) (float64, error) {
	args := m.Called(ctx, branchID)
	return args.Get(0).(float64), args.Error(1)
}

type MockGoldPriceRepository struct {
	mock.Mock
}

func (m *MockGoldPriceRepository) Create(ctx context.Context, price *entity.GoldPrice) error {
	args := m.Called(ctx, price)
	return args.Error(0)
}

func (m *MockGoldPriceRepository) GetCurrent(ctx context.Context) (*entity.GoldPrice, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.GoldPrice), args.Error(1)
}

func (m *MockGoldPriceRepository) GetHistory(ctx context.Context, limit int) ([]*entity.GoldPrice, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*entity.GoldPrice), args.Error(1)
}

func (m *MockGoldPriceRepository) GetByDateRange(ctx context.Context, from, to string) ([]*entity.GoldPrice, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).([]*entity.GoldPrice), args.Error(1)
}

func (m *MockGoldPriceRepository) DeactivateAll(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockProductItemRepository struct {
	mock.Mock
}

func (m *MockProductItemRepository) Create(ctx context.Context, item *entity.ProductItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockProductItemRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.ProductItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ProductItem), args.Error(1)
}

func (m *MockProductItemRepository) GetByBarcode(ctx context.Context, barcode string) (*entity.ProductItem, error) {
	args := m.Called(ctx, barcode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ProductItem), args.Error(1)
}

func (m *MockProductItemRepository) GetByProductID(ctx context.Context, productID primitive.ObjectID, status []entity.ProductStatus) ([]*entity.ProductItem, error) {
	args := m.Called(ctx, productID, status)
	return args.Get(0).([]*entity.ProductItem), args.Error(1)
}

func (m *MockProductItemRepository) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.ProductStatus) ([]*entity.ProductItem, error) {
	args := m.Called(ctx, branchID, status)
	return args.Get(0).([]*entity.ProductItem), args.Error(1)
}

func (m *MockProductItemRepository) Update(ctx context.Context, item *entity.ProductItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockProductItemRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockStockLogRepository struct {
	mock.Mock
}

func (m *MockStockLogRepository) Create(ctx context.Context, log *entity.StockLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockStockLogRepository) GetByProductID(ctx context.Context, productID primitive.ObjectID) ([]*entity.StockLog, error) {
	args := m.Called(ctx, productID)
	return args.Get(0).([]*entity.StockLog), args.Error(1)
}

func (m *MockStockLogRepository) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, limit, offset int) ([]*entity.StockLog, error) {
	args := m.Called(ctx, branchID, limit, offset)
	return args.Get(0).([]*entity.StockLog), args.Error(1)
}
