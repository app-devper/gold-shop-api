package sale

import (
	"context"
	"errors"

	"github.com/devper-gold/gold-shop-api/internal/domain/entity"
	"github.com/devper-gold/gold-shop-api/internal/domain/repository"
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

// Create creates a new sale
func (s *Service) Create(ctx context.Context, input CreateSaleInput) (*entity.Sale, error) {
	// Get branch for sale number generation
	branch, err := s.branchRepo.GetByID(ctx, input.BranchID)
	if err != nil || branch == nil {
		return nil, errors.New("branch not found")
	}

	// Generate sale number
	saleNumber, err := s.saleRepo.GenerateSaleNumber(ctx, branch.Code)
	if err != nil {
		return nil, errors.New("failed to generate sale number")
	}

	// Create sale
	sale := entity.NewSale(input.BranchID, input.UserID, saleNumber, input.SaleType)
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

	// Process items
	for _, itemInput := range input.Items {
		productID, err := primitive.ObjectIDFromHex(itemInput.ProductID)
		if err != nil {
			return nil, errors.New("invalid product ID")
		}

		product, err := s.productRepo.GetByID(ctx, productID)
		if err != nil || product == nil {
			return nil, errors.New("product not found")
		}

		// Validation
		if product.BranchID != input.BranchID {
			return nil, errors.New("product does not belong to this branch")
		}

		var unitPrice float64
		var laborCost float64
		var weight float64
		var productItemID *primitive.ObjectID

		if product.StockType == entity.StockTypePiece {
			if itemInput.ProductItemID == "" {
				return nil, errors.New("product item ID is required for piece-based products")
			}
			itemID, err := primitive.ObjectIDFromHex(itemInput.ProductItemID)
			if err != nil {
				return nil, errors.New("invalid product item ID")
			}
			productItemID = &itemID

			item, err := s.itemRepo.GetByID(ctx, itemID)
			if err != nil || item == nil {
				return nil, errors.New("product item not found")
			}
			if item.ProductID != productID {
				return nil, errors.New("item does not belong to the selected product")
			}
			if item.Status != entity.ProductStatusAvailable {
				return nil, errors.New("item is not available")
			}

			weight = item.Weight
			laborCost = item.LaborCost

			// Calculate dynamic price: (GoldOrnamentSell / 15.16) * weight
			gramPrice := goldPrice.GoldOrnamentSell / 15.16
			unitPrice = gramPrice * weight

			// Update item status
			item.Status = entity.ProductStatusSold
			if err := s.itemRepo.Update(ctx, item); err != nil {
				return nil, errors.New("failed to update item status")
			}
		} else {
			// Weight-based
			if itemInput.Weight <= 0 {
				return nil, errors.New("weight is required for weight-based products")
			}
			if product.Weight < itemInput.Weight {
				return nil, errors.New("insufficient weight in stock")
			}

			weight = itemInput.Weight
			laborCost = product.LaborCost // For weight-based, labor cost might be per item or per gram? Assume per gram if it's large, or per piece. Using as is.

			// Calculate dynamic price: (GoldBarSell / 15.244) * weight
			gramPrice := goldPrice.GoldBarSell / 15.244
			unitPrice = gramPrice * weight

			// Update product weight (stock)
			product.Weight -= weight
			if err := s.productRepo.Update(ctx, product); err != nil {
				return nil, errors.New("failed to update product stock")
			}
		}

		// Use manual price if provided (overwrites dynamic price)
		if itemInput.Price != nil {
			unitPrice = *itemInput.Price
		}

		total := unitPrice + laborCost

		// Apply item discount
		if itemInput.Discount > 0 {
			if itemInput.DiscountType == entity.DiscountTypePercent {
				total -= total * (itemInput.Discount / 100)
			} else {
				total -= itemInput.Discount
			}
		}

		saleItem := entity.SaleItem{
			ProductID:     productID,
			ProductItemID: productItemID,
			ProductName:   product.Name,
			GoldType:      product.GoldType,
			Weight:        weight,
			PriceLevel:    itemInput.PriceLevel,
			UnitPrice:     unitPrice,
			LaborCost:     laborCost,
			Discount:      itemInput.Discount,
			DiscountType:  itemInput.DiscountType,
			Cost:          product.Cost,
			Total:         total,
		}

		sale.AddItem(saleItem)

		// Record Stock Log
		stockLog := entity.NewStockLog(input.BranchID, productID, input.UserID, entity.StockActionSale, weight)
		stockLog.ProductItemID = productItemID
		stockLog.ReferenceID = saleNumber
		s.stockLogRepo.Create(ctx, stockLog)
	}

	// Process old gold items (for buy_old or exchange)
	for _, oldGold := range input.OldGoldItems {
		oldGoldItem := entity.OldGoldItem{
			Description:  oldGold.Description,
			GoldType:     oldGold.GoldType,
			Weight:       oldGold.Weight,
			PricePerUnit: oldGold.PricePerUnit,
			Total:        oldGold.Weight * oldGold.PricePerUnit,
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

	// Calculate points earned (1 point per 100 baht)
	sale.PointsEarned = int(sale.NetTotal / 100)

	// Update customer points and spending
	if sale.CustomerID != nil {
		customer, err := s.customerRepo.GetByID(ctx, *sale.CustomerID)
		if err == nil {
			if customer.IsMember {
				// Deduct used points
				if sale.PointsUsed > 0 {
					if !customer.RedeemPoints(sale.PointsUsed) {
						return nil, errors.New("insufficient points")
					}
				}
				// Add earned points
				customer.AddPoints(sale.PointsEarned)
			}
			// Update total spending
			customer.TotalSpent += sale.NetTotal
			s.customerRepo.Update(ctx, customer)
		}
	}

	// Mark as completed if fully paid
	if sale.IsFullyPaid() {
		sale.Complete()
	}

	// Save sale
	if err := s.saleRepo.Create(ctx, sale); err != nil {
		return nil, errors.New("failed to create sale")
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

	// Restore product status or weight
	for _, item := range sale.Items {
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err == nil {
			if item.ProductItemID != nil {
				// Piece-based
				productItem, err := s.itemRepo.GetByID(ctx, *item.ProductItemID)
				if err == nil && productItem != nil {
					productItem.Status = entity.ProductStatusAvailable
					s.itemRepo.Update(ctx, productItem)
				}
			} else {
				// Weight-based
				product.Weight += item.Weight
				s.productRepo.Update(ctx, product)
			}

			// Record Stock Log for Cancellation
			stockLog := entity.NewStockLog(sale.BranchID, item.ProductID, sale.UserID, entity.StockActionCancel, item.Weight)
			stockLog.ProductItemID = item.ProductItemID
			stockLog.ReferenceID = sale.SaleNumber
			s.stockLogRepo.Create(ctx, stockLog)
		}
	}

	// Restore customer points and spending if needed
	if sale.CustomerID != nil {
		customer, err := s.customerRepo.GetByID(ctx, *sale.CustomerID)
		if err == nil {
			if customer.IsMember {
				// Restore used points
				customer.AddPoints(sale.PointsUsed)
				// Remove earned points
				customer.RedeemPoints(sale.PointsEarned)
			}
			// Deduct spending
			customer.TotalSpent -= sale.NetTotal
			if customer.TotalSpent < 0 {
				customer.TotalSpent = 0
			}
			s.customerRepo.Update(ctx, customer)
		}
	}

	sale.Cancel()
	return s.saleRepo.Update(ctx, sale)
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
