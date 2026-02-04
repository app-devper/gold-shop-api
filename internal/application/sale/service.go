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
	customerRepo repository.CustomerRepository
	branchRepo   repository.BranchRepository
	userRepo     repository.UserRepository
}

// NewService creates a new Sale service
func NewService(
	saleRepo repository.SaleRepository,
	productRepo repository.ProductRepository,
	customerRepo repository.CustomerRepository,
	branchRepo repository.BranchRepository,
	userRepo repository.UserRepository,
) *Service {
	return &Service{
		saleRepo:     saleRepo,
		productRepo:  productRepo,
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
	ProductID    string
	PriceLevel   string
	Discount     float64
	DiscountType entity.DiscountType
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

		if product.Status != entity.ProductStatusAvailable {
			return nil, errors.New("product is not available")
		}

		priceLevel := itemInput.PriceLevel
		if priceLevel == "" {
			priceLevel = "A"
		}

		unitPrice := product.GetPriceByLevel(priceLevel)
		total := unitPrice + product.LaborCost

		// Apply item discount
		if itemInput.Discount > 0 {
			if itemInput.DiscountType == entity.DiscountTypePercent {
				total -= total * (itemInput.Discount / 100)
			} else {
				total -= itemInput.Discount
			}
		}

		saleItem := entity.SaleItem{
			ProductID:    productID,
			ProductName:  product.Name,
			GoldType:     product.GoldType,
			Weight:       product.Weight,
			PriceLevel:   priceLevel,
			UnitPrice:    unitPrice,
			LaborCost:    product.LaborCost,
			Discount:     itemInput.Discount,
			DiscountType: itemInput.DiscountType,
			Cost:         product.Cost,
			Total:        total,
		}

		sale.AddItem(saleItem)

		// Update product status
		product.Status = entity.ProductStatusSold
		if err := s.productRepo.Update(ctx, product); err != nil {
			return nil, errors.New("failed to update product status")
		}
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

	// Update customer points if member
	if sale.CustomerID != nil {
		customer, err := s.customerRepo.GetByID(ctx, *sale.CustomerID)
		if err == nil && customer.IsMember {
			// Deduct used points
			if sale.PointsUsed > 0 {
				if !customer.RedeemPoints(sale.PointsUsed) {
					return nil, errors.New("insufficient points")
				}
			}
			// Add earned points
			customer.AddPoints(sale.PointsEarned)
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

	// Restore product status
	for _, item := range sale.Items {
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err == nil {
			product.Status = entity.ProductStatusAvailable
			s.productRepo.Update(ctx, product)
		}
	}

	// Restore customer points if needed
	if sale.CustomerID != nil {
		customer, err := s.customerRepo.GetByID(ctx, *sale.CustomerID)
		if err == nil && customer.IsMember {
			// Restore used points
			customer.AddPoints(sale.PointsUsed)
			// Remove earned points
			customer.RedeemPoints(sale.PointsEarned)
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
