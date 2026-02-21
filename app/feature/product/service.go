package product

import (
	"context"
	"errors"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateProductRequest represents data for creating a product
type CreateProductRequest struct {
	BranchID   string           `json:"branch_id" binding:"required"`
	CategoryID string           `json:"category_id" binding:"required"`
	SKU        string           `json:"sku" binding:"required"`
	Name       string           `json:"name" binding:"required"`
	StockType  entity.StockType `json:"stock_type" binding:"required"`
	GoldType   string           `json:"gold_type" binding:"required"`
	Weight     float64          `json:"weight" binding:"required"`
	LaborCost  float64          `json:"labor_cost"`
	Cost       float64          `json:"cost"`
	Barcode    string           `json:"barcode"`
}

// UpdateProductRequest represents data for updating a product
type UpdateProductRequest struct {
	Name      string  `json:"name"`
	GoldType  string  `json:"gold_type"`
	Weight    float64 `json:"weight"`
	LaborCost float64 `json:"labor_cost"`
	Cost      float64 `json:"cost"`
	Price     float64 `json:"price"`
	Barcode   string  `json:"barcode"`
	Status    string  `json:"status"`
}

// Service handles product business logic
type Service struct {
	productRepo  repository.ProductRepository
	itemRepo     repository.ProductItemRepository
	stockLogRepo repository.StockLogRepository
	categoryRepo repository.ProductCategoryRepository
	branchRepo   repository.BranchRepository
}

// NewService creates a new product service
func NewService(
	productRepo repository.ProductRepository,
	itemRepo repository.ProductItemRepository,
	stockLogRepo repository.StockLogRepository,
	categoryRepo repository.ProductCategoryRepository,
	branchRepo repository.BranchRepository,
) *Service {
	return &Service{
		productRepo:  productRepo,
		itemRepo:     itemRepo,
		stockLogRepo: stockLogRepo,
		categoryRepo: categoryRepo,
		branchRepo:   branchRepo,
	}
}

// CreateProduct creates a new product
func (s *Service) CreateProduct(ctx context.Context, req *CreateProductRequest) (*entity.Product, error) {
	branchID, err := primitive.ObjectIDFromHex(req.BranchID)
	if err != nil {
		return nil, errors.New("invalid branch ID")
	}

	categoryID, err := primitive.ObjectIDFromHex(req.CategoryID)
	if err != nil {
		return nil, errors.New("invalid category ID")
	}

	// Verify branch and category exist
	if _, err := s.branchRepo.GetByID(ctx, branchID); err != nil {
		return nil, errors.New("branch not found")
	}
	if _, err := s.categoryRepo.GetByID(ctx, categoryID); err != nil {
		return nil, errors.New("category not found")
	}

	// Check redundant SKU?
	existing, _ := s.productRepo.GetBySKU(ctx, req.SKU)
	if existing != nil {
		return nil, errors.New("SKU already exists")
	}

	product := entity.NewProduct(branchID, categoryID, req.SKU, req.Name, req.GoldType, req.StockType, req.Weight, req.LaborCost)
	product.Cost = req.Cost
	product.Barcode = req.Barcode

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

// GetProduct retrieves a product by ID
func (s *Service) GetProduct(ctx context.Context, id string) (*entity.Product, error) {
	productID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid product ID")
	}

	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil || product == nil {
		return product, err
	}

	if product.DefaultStockType() == entity.StockTypePiece {
		items, err := s.itemRepo.GetByProductID(ctx, product.ID, []entity.ProductStatus{entity.ProductStatusAvailable})
		if err != nil {
			return nil, err
		}
		product.Items = items
	}

	return product, nil
}

// GetProducts retrieves products with filtering
func (s *Service) GetProducts(ctx context.Context, branchID string, status string, limit, offset int) ([]*entity.Product, error) {
	bID, err := primitive.ObjectIDFromHex(branchID)
	if err != nil {
		return nil, errors.New("invalid branch ID")
	}

	var statuses []entity.ProductStatus
	if status != "" && status != "all" {
		statuses = []entity.ProductStatus{entity.ProductStatus(status)}
	}

	products, err := s.productRepo.GetByBranchID(ctx, bID, statuses, limit, offset)
	if err != nil {
		return nil, err
	}

	for _, p := range products {
		if p.DefaultStockType() == entity.StockTypePiece {
			items, err := s.itemRepo.GetByProductID(ctx, p.ID, []entity.ProductStatus{entity.ProductStatusAvailable})
			if err != nil {
				return nil, err
			}
			p.Items = items
		}
	}

	return products, nil
}

// UpdateProduct updates a product
func (s *Service) UpdateProduct(ctx context.Context, id string, req *UpdateProductRequest) (*entity.Product, error) {
	productID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid product ID")
	}

	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.GoldType != "" {
		product.GoldType = req.GoldType
	}
	if req.Weight > 0 {
		product.Weight = req.Weight
	}
	// LaborCost and Cost can be 0, so check logic might need to rely on pointer-based struct or separate boolean flags if we want to allow updating to 0. For now assuming > 0 or update if field is present in JSON map which manual unmarshal would handle. But here simplified.
	product.LaborCost = req.LaborCost
	product.Cost = req.Cost

	product.Price = req.Price
	if req.Barcode != "" {
		product.Barcode = req.Barcode
	}
	if req.Status != "" {
		product.Status = entity.ProductStatus(req.Status)
	}

	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

// DeleteProduct deletes a product
func (s *Service) DeleteProduct(ctx context.Context, id string) error {
	productID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid product ID")
	}

	return s.productRepo.Delete(ctx, productID)
}

// CreateCategory creates a new product category
func (s *Service) CreateCategory(ctx context.Context, name, code, description string) (*entity.ProductCategory, error) {
	category := &entity.ProductCategory{
		Name:        name,
		Code:        code,
		Description: description,
		IsActive:    true,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// GetCategories retrieves all product categories
func (s *Service) GetCategories(ctx context.Context) ([]*entity.ProductCategory, error) {
	return s.categoryRepo.GetAll(ctx)
}
