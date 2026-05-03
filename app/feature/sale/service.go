package sale

import (
	"context"
	"errors"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service handles sale logic
type Service struct {
	saleRepo     repository.SaleRepository
	productRepo  repository.ProductRepository
	itemRepo     repository.ProductItemRepository
	priceRepo    repository.GoldPriceRepository
	stockLogRepo repository.StockLogRepository
	customerRepo repository.CustomerRepository
	branchRepo   repository.BranchRepository
	userRepo     repository.UserRepository
	txManager    repository.TransactionManager
}

// NewService creates a new Sale service
func NewService(
	saleRepo repository.SaleRepository,
	productRepo repository.ProductRepository,
	itemRepo repository.ProductItemRepository,
	priceRepo repository.GoldPriceRepository,
	stockLogRepo repository.StockLogRepository,
	customerRepo repository.CustomerRepository,
	branchRepo repository.BranchRepository,
	userRepo repository.UserRepository,
	txManager repository.TransactionManager,
) *Service {
	return &Service{
		saleRepo:     saleRepo,
		productRepo:  productRepo,
		itemRepo:     itemRepo,
		priceRepo:    priceRepo,
		stockLogRepo: stockLogRepo,
		customerRepo: customerRepo,
		branchRepo:   branchRepo,
		userRepo:     userRepo,
		txManager:    txManager,
	}
}

// CreateSaleInput represents input for creating a sale
type CreateSaleInput struct {
	BranchID     primitive.ObjectID
	UserID       primitive.ObjectID
	CustomerID   string
	SaleType     entity.SaleType
	Items        []SaleItemInput
	OldGoldItems []OldGoldInput
	Discount     float64
	DiscountType entity.DiscountType
	Payments     []PaymentInput
	PointsUsed   int
	Notes        string
}

// SaleItemInput represents input for a sale item
type SaleItemInput struct {
	ProductID     string
	ProductItemID string
	PriceLevel    string
	Price         *float64
	Weight        float64
	Discount      float64
	DiscountType  entity.DiscountType
}

// OldGoldInput represents input for old gold
type OldGoldInput struct {
	Description  string
	GoldType     string
	Weight       float64
	PricePerUnit float64
}

// PaymentInput represents input for a payment
type PaymentInput struct {
	Method    entity.PaymentMethod
	Amount    float64
	Reference string
}

// resolvedItem holds validated data for a single sale line item (Phase 1 output)
type resolvedItem struct {
	saleItem      entity.SaleItem
	product       *entity.Product
	productItem   *entity.ProductItem // non-nil for piece-based
	productItemID *primitive.ObjectID
	weight        float64
}

// calculateGramPrice returns the per-gram sell price based on the product's gold type
func calculateGramPrice(goldPrice *entity.GoldPrice, product *entity.Product) float64 {
	if product.IsBarGold() {
		return goldPrice.GoldBarSell / entity.BahtPerGramBar
	}
	return goldPrice.GoldOrnamentSell / entity.BahtPerGramOrnament
}

// Create creates a new sale using a two-phase approach:
// Phase 1: Validate all inputs and build sale items (no DB mutations)
// Phase 2: Apply all DB mutations (stock updates, stock logs, customer updates)
func (s *Service) Create(ctx context.Context, input CreateSaleInput) (*entity.Sale, error) {
	// Get branch for sale number generation
	branch, err := s.branchRepo.GetByID(ctx, input.BranchID)
	if err != nil || branch == nil {
		return nil, errors.New("branch not found")
	}

	if input.Discount < 0 {
		return nil, errors.New("discount cannot be negative")
	}

	// Create sale (sale number will be generated inside the transaction)
	sale := entity.NewSale(input.BranchID, input.UserID, "", input.SaleType)
	sale.Discount = input.Discount
	sale.DiscountType = input.DiscountType
	sale.PointsUsed = input.PointsUsed
	sale.Notes = input.Notes

	// Set customer if provided
	if input.CustomerID != "" {
		customerID, err := primitive.ObjectIDFromHex(input.CustomerID)
		if err == nil {
			sale.CustomerID = &customerID
		}
	}

	// Get current gold price for dynamic pricing
	goldPrice, err := s.priceRepo.GetCurrent(ctx)
	if err != nil || goldPrice == nil {
		return nil, errors.New("could not fetch current gold price")
	}

	// ── Phase 1: Validate all items (no DB writes) ──
	resolved := make([]resolvedItem, 0, len(input.Items))

	for _, itemInput := range input.Items {
		productID, err := primitive.ObjectIDFromHex(itemInput.ProductID)
		if err != nil {
			return nil, errors.New("invalid product ID")
		}

		product, err := s.productRepo.GetByID(ctx, productID)
		if err != nil || product == nil {
			return nil, errors.New("product not found")
		}

		if product.BranchID != input.BranchID {
			return nil, errors.New("product does not belong to this branch")
		}
		if !product.IsActive {
			return nil, errors.New("product is no longer active")
		}

		// Every product is piece-based — operator must specify the exact item.
		if itemInput.ProductItemID == "" {
			return nil, errors.New("product item ID is required")
		}
		itemID, err := primitive.ObjectIDFromHex(itemInput.ProductItemID)
		if err != nil {
			return nil, errors.New("invalid product item ID")
		}
		productItemID := &itemID

		productItem, err := s.itemRepo.GetByID(ctx, itemID)
		if err != nil || productItem == nil {
			return nil, errors.New("product item not found")
		}
		if productItem.ProductID != productID {
			return nil, errors.New("item does not belong to the selected product")
		}
		if productItem.Status != entity.ProductStatusAvailable {
			return nil, errors.New("item is not available")
		}

		weight := productItem.WeightGrams
		laborCost := productItem.LaborCost
		gramPrice := calculateGramPrice(goldPrice, product)
		unitPrice := gramPrice * weight

		// Use manual price if provided (overwrites dynamic price)
		if itemInput.Price != nil {
			unitPrice = *itemInput.Price
		}

		if itemInput.Discount < 0 {
			return nil, errors.New("item discount cannot be negative")
		}

		total := unitPrice + laborCost
		if itemInput.Discount > 0 {
			if itemInput.DiscountType == entity.DiscountTypePercent {
				total -= total * (itemInput.Discount / 100)
			} else {
				total -= itemInput.Discount
			}
		}
		if total < 0 {
			total = 0
		}
		total = utils.RoundBaht(total)

		saleItem := entity.SaleItem{
			ProductID:     productID,
			ProductItemID: productItemID,
			ProductName:   product.Name,
			GoldType:      product.GoldType,
			Weight:        weight,
			PriceLevel:    itemInput.PriceLevel,
			UnitPrice:     utils.RoundBaht(unitPrice),
			LaborCost:     utils.RoundBaht(laborCost),
			Discount:      itemInput.Discount,
			DiscountType:  itemInput.DiscountType,
			Cost:          productItem.Cost,
			Total:         total,
		}

		resolved = append(resolved, resolvedItem{
			saleItem:      saleItem,
			product:       product,
			productItem:   productItem,
			productItemID: productItemID,
			weight:        weight,
		})
	}

	// Process old gold items (for buy_old or exchange)
	for _, oldGold := range input.OldGoldItems {
		oldGoldItem := entity.OldGoldItem{
			Description:  oldGold.Description,
			GoldType:     oldGold.GoldType,
			Weight:       oldGold.Weight,
			PricePerUnit: oldGold.PricePerUnit,
			Total:        utils.RoundBaht(oldGold.Weight * oldGold.PricePerUnit),
		}
		sale.AddOldGoldItem(oldGoldItem)
	}

	// Process payments
	for _, paymentInput := range input.Payments {
		payment := entity.Payment{
			Method:    paymentInput.Method,
			Amount:    paymentInput.Amount,
			Reference: paymentInput.Reference,
		}
		sale.AddPayment(payment)
	}

	// Add validated items to sale (builds totals)
	for _, r := range resolved {
		sale.AddItem(r.saleItem)
	}

	// Calculate points earned (1 point per 100 baht, never negative)
	if sale.NetTotal > 0 {
		sale.PointsEarned = int(sale.NetTotal / 100)
	} else {
		sale.PointsEarned = 0
	}

	// Validate customer points before any mutations
	if sale.CustomerID != nil && sale.PointsUsed > 0 {
		customer, err := s.customerRepo.GetByID(ctx, *sale.CustomerID)
		if err != nil {
			return nil, errors.New("customer not found")
		}
		if !customer.IsMember || customer.Membership == nil || customer.Membership.Points < sale.PointsUsed {
			return nil, errors.New("insufficient points")
		}
	}

	// Mark as completed if fully paid
	if sale.IsFullyPaid() {
		sale.Complete()
	}

	// ── Phase 2: Apply all DB mutations inside a transaction ──
	txErr := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Generate sale number inside the transaction so it rolls back on failure
		saleNumber, err := s.saleRepo.GenerateSaleNumber(txCtx, branch.Code)
		if err != nil {
			return errors.New("failed to generate sale number")
		}
		sale.SaleNumber = saleNumber

		// Save sale record (status already set)
		if err := s.saleRepo.Create(txCtx, sale); err != nil {
			return errors.New("failed to create sale")
		}

		for _, r := range resolved {
			r.productItem.Status = entity.ProductStatusSold
			if err := s.itemRepo.Update(txCtx, r.productItem); err != nil {
				return errors.New("failed to update item status")
			}

			stockLog := entity.NewStockLog(input.BranchID, r.saleItem.ProductID, input.UserID, entity.StockActionSale, r.weight)
			stockLog.ProductItemID = r.productItemID
			stockLog.ReferenceID = saleNumber
			if err := s.stockLogRepo.Create(txCtx, stockLog); err != nil {
				return errors.New("failed to record stock log")
			}
		}

		if sale.CustomerID != nil {
			customer, err := s.customerRepo.GetByID(txCtx, *sale.CustomerID)
			if err != nil {
				return errors.New("failed to fetch customer")
			}
			if customer != nil {
				if customer.IsMember {
					if sale.PointsUsed > 0 {
						if !customer.RedeemPoints(sale.PointsUsed) {
							return errors.New("insufficient points")
						}
					}
					customer.AddPoints(sale.PointsEarned)
				}
				customer.TotalSpent += sale.NetTotal
				if err := s.customerRepo.Update(txCtx, customer); err != nil {
					return errors.New("failed to update customer")
				}
			}
		}

		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return sale, nil
}

// GetByID retrieves a sale by ID
func (s *Service) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Sale, error) {
	return s.saleRepo.GetByID(ctx, id)
}

// GetByBranchID retrieves sales by branch ID
func (s *Service) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.SaleStatus, limit, offset int) ([]*entity.Sale, error) {
	return s.saleRepo.GetByBranchID(ctx, branchID, status, limit, offset)
}

// GetUnpaidSales retrieves sales that are not fully paid
func (s *Service) GetUnpaidSales(ctx context.Context, branchID primitive.ObjectID) ([]*entity.Sale, error) {
	return s.saleRepo.GetUnpaidByBranchID(ctx, branchID)
}

// Cancel cancels a sale
func (s *Service) Cancel(ctx context.Context, id primitive.ObjectID) error {
	sale, err := s.saleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if sale.Status == entity.SaleStatusCancelled {
		return errors.New("sale is already cancelled")
	}
	if sale.Status == entity.SaleStatusCompleted {
		return errors.New("cannot cancel a completed sale")
	}

	sale.Cancel()

	return s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Restore each item to available; product master holds no balance to roll back.
		for _, item := range sale.Items {
			if item.ProductItemID == nil {
				return errors.New("legacy weight-based sale cannot be cancelled in this version")
			}
			productItem, err := s.itemRepo.GetByID(txCtx, *item.ProductItemID)
			if err != nil || productItem == nil {
				return errors.New("failed to find product item for cancellation")
			}
			productItem.Status = entity.ProductStatusAvailable
			if err := s.itemRepo.Update(txCtx, productItem); err != nil {
				return errors.New("failed to restore product item status")
			}

			stockLog := entity.NewStockLog(sale.BranchID, item.ProductID, sale.UserID, entity.StockActionCancel, item.Weight)
			stockLog.ProductItemID = item.ProductItemID
			stockLog.ReferenceID = sale.SaleNumber
			if err := s.stockLogRepo.Create(txCtx, stockLog); err != nil {
				return errors.New("failed to record cancellation stock log")
			}
		}

		// Restore customer points and spending if needed
		if sale.CustomerID != nil {
			customer, err := s.customerRepo.GetByID(txCtx, *sale.CustomerID)
			if err != nil {
				return errors.New("failed to fetch customer for cancellation")
			}
			if customer != nil {
				if customer.IsMember {
					// Restore used points
					customer.AddPoints(sale.PointsUsed)
					// Remove earned points — if customer already spent them, cap at 0
					if customer.Membership != nil && customer.Membership.Points >= sale.PointsEarned {
						customer.RedeemPoints(sale.PointsEarned)
					} else if customer.Membership != nil {
						customer.Membership.Points = 0
					}
				}
				// Deduct spending
				customer.TotalSpent -= sale.NetTotal
				if customer.TotalSpent < 0 {
					customer.TotalSpent = 0
				}
				if err := s.customerRepo.Update(txCtx, customer); err != nil {
					return errors.New("failed to update customer on cancellation")
				}
			}
		}

		return s.saleRepo.Update(txCtx, sale)
	})
}

// Receipt represents a sale receipt
type Receipt struct {
	Sale        *entity.Sale `json:"sale"`
	BranchName  string       `json:"branch_name"`
	CashierName string       `json:"cashier_name"`
}

// GenerateReceipt generates a receipt for a sale
func (s *Service) GenerateReceipt(ctx context.Context, saleID primitive.ObjectID) (*Receipt, error) {
	sale, err := s.saleRepo.GetByID(ctx, saleID)
	if err != nil || sale == nil {
		return nil, errors.New("sale not found")
	}

	branch, _ := s.branchRepo.GetByID(ctx, sale.BranchID)
	branchName := ""
	if branch != nil {
		branchName = branch.Name
	}

	cashierName := ""
	user, _ := s.userRepo.GetByID(ctx, sale.UserID)
	if user != nil {
		cashierName = user.FullName
	}

	return &Receipt{
		Sale:        sale,
		BranchName:  branchName,
		CashierName: cashierName,
	}, nil
}
