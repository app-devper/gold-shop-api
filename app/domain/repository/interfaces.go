package repository

import (
	"context"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TransactionManager provides database transaction support
type TransactionManager interface {
	// WithTransaction executes fn inside an atomic transaction.
	// If fn returns an error the transaction is aborted; otherwise it is committed.
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// EmployeeRepository defines employee data operations
type EmployeeRepository interface {
	Create(ctx context.Context, employee *entity.Employee) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Employee, error)
	GetByUserID(ctx context.Context, userID string) (*entity.Employee, error)
	GetAll(ctx context.Context) ([]*entity.Employee, error)
	GetByBranchID(ctx context.Context, branchID primitive.ObjectID) ([]*entity.Employee, error)
	Update(ctx context.Context, employee *entity.Employee) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

// BranchRepository defines branch data operations
type BranchRepository interface {
	Create(ctx context.Context, branch *entity.Branch) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Branch, error)
	GetByCode(ctx context.Context, code string) (*entity.Branch, error)
	GetAll(ctx context.Context) ([]*entity.Branch, error)
	Update(ctx context.Context, branch *entity.Branch) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

// UserRepository defines user data operations
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	GetByBranchID(ctx context.Context, branchID primitive.ObjectID) ([]*entity.User, error)
	GetAll(ctx context.Context) ([]*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

// CustomerRepository defines customer data operations
type CustomerRepository interface {
	Create(ctx context.Context, customer *entity.Customer) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Customer, error)
	GetByMemberCode(ctx context.Context, memberCode string) (*entity.Customer, error)
	GetByRFID(ctx context.Context, rfidCard string) (*entity.Customer, error)
	GetByPhone(ctx context.Context, phone string) (*entity.Customer, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entity.Customer, error)
	Search(ctx context.Context, query string, limit int) ([]*entity.Customer, error)
	Update(ctx context.Context, customer *entity.Customer) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
}

// ProductRepository defines product (catalog) data operations.
// Stock is held entirely on ProductItems; the master never carries weight/price.
type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Product, error)
	GetBySKU(ctx context.Context, sku string) (*entity.Product, error)
	GetByBranchID(ctx context.Context, branchID primitive.ObjectID, kind entity.ProductKind, search string, limit, offset int) ([]*entity.Product, error)
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context, branchID primitive.ObjectID) (int64, error)
}

// ProductItemRepository defines individual item data operations
type ProductItemRepository interface {
	Create(ctx context.Context, item *entity.ProductItem) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*entity.ProductItem, error)
	GetByBarcode(ctx context.Context, barcode string) (*entity.ProductItem, error)
	GetByProductID(ctx context.Context, productID primitive.ObjectID, status []entity.ProductStatus) ([]*entity.ProductItem, error)
	GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.ProductStatus) ([]*entity.ProductItem, error)
	Update(ctx context.Context, item *entity.ProductItem) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

// StockLogRepository defines stock log operations
type StockLogRepository interface {
	Create(ctx context.Context, log *entity.StockLog) error
	GetByProductID(ctx context.Context, productID primitive.ObjectID) ([]*entity.StockLog, error)
	GetByBranchID(ctx context.Context, branchID primitive.ObjectID, limit, offset int) ([]*entity.StockLog, error)
}

// GoldPriceRepository defines gold price data operations
type GoldPriceRepository interface {
	Create(ctx context.Context, price *entity.GoldPrice) error
	GetCurrent(ctx context.Context) (*entity.GoldPrice, error)
	GetHistory(ctx context.Context, limit int) ([]*entity.GoldPrice, error)
	GetByDateRange(ctx context.Context, from, to string) ([]*entity.GoldPrice, error)
	DeactivateAll(ctx context.Context) error
}

// SaleRepository defines sales data operations
type SaleRepository interface {
	Create(ctx context.Context, sale *entity.Sale) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Sale, error)
	GetBySaleNumber(ctx context.Context, saleNumber string) (*entity.Sale, error)
	GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.SaleStatus, limit, offset int) ([]*entity.Sale, error)
	GetByCustomerID(ctx context.Context, customerID primitive.ObjectID, limit int) ([]*entity.Sale, error)
	GetByDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) ([]*entity.Sale, error)
	Update(ctx context.Context, sale *entity.Sale) error
	GenerateSaleNumber(ctx context.Context, branchCode string, saleType entity.SaleType) (string, error)
	SumByBranchAndDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) (float64, error)
	CountByBranchAndDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) (int64, error)
	SumCostByBranchAndDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) (float64, error)
	GetUnpaidByBranchID(ctx context.Context, branchID primitive.ObjectID) ([]*entity.Sale, error)
	GetTopSellingProducts(ctx context.Context, branchID primitive.ObjectID, from, to string, limit int) ([]TopProduct, error)
	GetEmployeePerformance(ctx context.Context, branchID primitive.ObjectID, from, to string) ([]EmployeePerformance, error)
	GetSalesTrends(ctx context.Context, branchID primitive.ObjectID, from, to string) ([]SalesTrend, error)
}

// TopProduct represents a top selling product
type TopProduct struct {
	ProductID   primitive.ObjectID `json:"product_id" bson:"_id"`
	ProductName string             `json:"product_name" bson:"product_name"`
	TotalQty    int                `json:"total_qty" bson:"total_qty"`
	TotalRev    float64            `json:"total_rev" bson:"total_rev"`
}

// EmployeePerformance represents sales performance of an employee
type EmployeePerformance struct {
	UserID       primitive.ObjectID `json:"user_id" bson:"_id"`
	FullName     string             `json:"full_name" bson:"full_name"`
	TotalSales   float64            `json:"total_sales" bson:"total_sales"`
	SaleCount    int                `json:"sale_count" bson:"sale_count"`
	AvgSaleValue float64            `json:"avg_sale_value" bson:"avg_sale_value"`
}

// SalesTrend represents daily sales trend
type SalesTrend struct {
	Date      string  `json:"date" bson:"_id"`
	Revenue   float64 `json:"revenue" bson:"revenue"`
	Cost      float64 `json:"cost" bson:"cost"`
	Profit    float64 `json:"profit" bson:"profit"`
	SaleCount int     `json:"sale_count" bson:"sale_count"`
}

// PawnRepository defines pawn data operations
type PawnRepository interface {
	Create(ctx context.Context, pawn *entity.Pawn) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Pawn, error)
	GetByPawnNumber(ctx context.Context, pawnNumber string) (*entity.Pawn, error)
	GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.PawnStatus, limit, offset int) ([]*entity.Pawn, error)
	GetByCustomerID(ctx context.Context, customerID primitive.ObjectID) ([]*entity.Pawn, error)
	GetDueSoon(ctx context.Context, branchID primitive.ObjectID, days int) ([]*entity.Pawn, error)
	GetOverdue(ctx context.Context, branchID primitive.ObjectID) ([]*entity.Pawn, error)
	Update(ctx context.Context, pawn *entity.Pawn) error
	GeneratePawnNumber(ctx context.Context, branchCode string) (string, error)
	CountByStatus(ctx context.Context, branchID primitive.ObjectID, status entity.PawnStatus) (int64, error)
	SumInterestByBranchAndDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) (float64, error)
}

// GoldSavingRepository defines gold saving data operations
type GoldSavingRepository interface {
	Create(ctx context.Context, goldSaving *entity.GoldSaving) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*entity.GoldSaving, error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (*entity.GoldSaving, error)
	GetByCustomerID(ctx context.Context, customerID primitive.ObjectID) ([]*entity.GoldSaving, error)
	GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.GoldSavingStatus) ([]*entity.GoldSaving, error)
	Update(ctx context.Context, goldSaving *entity.GoldSaving) error
	GenerateAccountNumber(ctx context.Context, branchCode string) (string, error)
	SumBalanceByBranch(ctx context.Context, branchID primitive.ObjectID) (float64, error)
}

// ExpenseCategoryRepository defines expense category operations
type ExpenseCategoryRepository interface {
	Create(ctx context.Context, category *entity.ExpenseCategory) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*entity.ExpenseCategory, error)
	GetAll(ctx context.Context) ([]*entity.ExpenseCategory, error)
	Update(ctx context.Context, category *entity.ExpenseCategory) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

// ExpenseRepository defines expense data operations
type ExpenseRepository interface {
	Create(ctx context.Context, expense *entity.Expense) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Expense, error)
	GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.ExpenseStatus, limit, offset int) ([]*entity.Expense, error)
	GetByDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) ([]*entity.Expense, error)
	GetByCategoryID(ctx context.Context, categoryID primitive.ObjectID) ([]*entity.Expense, error)
	Update(ctx context.Context, expense *entity.Expense) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	SumByBranchAndDateRange(ctx context.Context, branchID primitive.ObjectID, from, to string) (float64, error)
}

// InventoryTransferRepository defines inventory transfer operations
type InventoryTransferRepository interface {
	Create(ctx context.Context, transfer *entity.InventoryTransfer) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*entity.InventoryTransfer, error)
	GetByTransferNumber(ctx context.Context, transferNumber string) (*entity.InventoryTransfer, error)
	GetByFromBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.TransferStatus) ([]*entity.InventoryTransfer, error)
	GetByToBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.TransferStatus) ([]*entity.InventoryTransfer, error)
	Update(ctx context.Context, transfer *entity.InventoryTransfer) error
	GenerateTransferNumber(ctx context.Context) (string, error)
}

// RewardRepository defines reward data operations
type RewardRepository interface {
	Create(ctx context.Context, reward *entity.Reward) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Reward, error)
	GetAll(ctx context.Context, activeOnly bool) ([]*entity.Reward, error)
	Update(ctx context.Context, reward *entity.Reward) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

// RewardRedemptionRepository defines reward redemption operations
type RewardRedemptionRepository interface {
	Create(ctx context.Context, redemption *entity.RewardRedemption) error
	GetByCustomerID(ctx context.Context, customerID primitive.ObjectID) ([]*entity.RewardRedemption, error)
	GetByBranchID(ctx context.Context, branchID primitive.ObjectID, limit, offset int) ([]*entity.RewardRedemption, error)
}
