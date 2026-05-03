package product

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service handles product (catalog) and product-item (physical stock) operations.
//
// Catalog masters carry no balance and no price; every sellable unit is a
// ProductItem keyed by Barcode. Two catalog kinds:
//
//	ornament — Design (free text), DefaultLaborCost; per-item weight is entered
//	           manually because every piece is unique.
//	bar      — BarSizeBaht (free input). New items default to BarSizeBaht ×
//	           BahtPerGramBar (15.244 g/baht); operator may override per item.
type Service struct {
	productRepo  repository.ProductRepository
	itemRepo     repository.ProductItemRepository
	stockLogRepo repository.StockLogRepository
	branchRepo   repository.BranchRepository
}

func NewService(
	productRepo repository.ProductRepository,
	itemRepo repository.ProductItemRepository,
	stockLogRepo repository.StockLogRepository,
	branchRepo repository.BranchRepository,
) *Service {
	return &Service{
		productRepo:  productRepo,
		itemRepo:     itemRepo,
		stockLogRepo: stockLogRepo,
		branchRepo:   branchRepo,
	}
}

// ── Master CRUD ───────────────────────────────────────────────────────────────

// CreateProductInput is shape-checked against Kind: ornament needs Design, bar
// needs BarSizeBaht > 0.
type CreateProductInput struct {
	BranchID         primitive.ObjectID
	SKU              string
	Code             string
	Kind             entity.ProductKind
	GoldType         string
	Name             string
	Description      string
	Design           string   // ornament-only
	BarSizeBaht      *float64 // bar-only
	DefaultLaborCost float64  // ornament-only
	Images           []string
}

func (s *Service) CreateProduct(ctx context.Context, in CreateProductInput) (*entity.Product, error) {
	if err := validateCreate(&in); err != nil {
		return nil, err
	}

	if existing, _ := s.productRepo.GetBySKU(ctx, in.SKU); existing != nil {
		return nil, errors.New("SKU already exists")
	}

	var product *entity.Product
	switch in.Kind {
	case entity.KindOrnament:
		product = entity.NewOrnamentProduct(in.BranchID, in.SKU, in.Name, in.Design, in.GoldType, in.DefaultLaborCost)
	case entity.KindBar:
		product = entity.NewBarProduct(in.BranchID, in.SKU, in.Name, in.GoldType, *in.BarSizeBaht)
	}
	product.Code = in.Code
	product.Description = in.Description
	if in.Images != nil {
		product.Images = in.Images
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

// UpdateProductInput updates display fields on the master.
// Kind, SKU and BranchID are immutable after creation — change requires creating a new product.
type UpdateProductInput struct {
	Name             *string
	Description      *string
	Design           *string
	DefaultLaborCost *float64
	BarSizeBaht      *float64
	Images           []string
	IsActive         *bool
}

func (s *Service) UpdateProduct(ctx context.Context, id primitive.ObjectID, in UpdateProductInput) (*entity.Product, error) {
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil || product == nil {
		return nil, errors.New("product not found")
	}
	if in.Name != nil {
		product.Name = *in.Name
	}
	if in.Description != nil {
		product.Description = *in.Description
	}
	if product.Kind == entity.KindOrnament {
		if in.Design != nil {
			product.Design = *in.Design
		}
		if in.DefaultLaborCost != nil {
			product.DefaultLaborCost = *in.DefaultLaborCost
		}
	}
	if product.Kind == entity.KindBar && in.BarSizeBaht != nil && *in.BarSizeBaht > 0 {
		size := *in.BarSizeBaht
		product.BarSizeBaht = &size
	}
	if in.Images != nil {
		product.Images = in.Images
	}
	if in.IsActive != nil {
		product.IsActive = *in.IsActive
	}

	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

func (s *Service) DeleteProduct(ctx context.Context, id primitive.ObjectID) error {
	return s.productRepo.Delete(ctx, id)
}

// GetProduct returns the master with its currently-available items populated.
func (s *Service) GetProduct(ctx context.Context, id primitive.ObjectID) (*entity.Product, error) {
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil || product == nil {
		return nil, err
	}
	items, _ := s.itemRepo.GetByProductID(ctx, product.ID, []entity.ProductStatus{entity.ProductStatusAvailable})
	product.Items = items
	return product, nil
}

// ListProductsInput filters at the catalog level.
type ListProductsInput struct {
	BranchID primitive.ObjectID
	Kind     entity.ProductKind // "" = all
	Search   string             // free text against name/sku/design
	Limit    int
	Offset   int
}

// ListProducts returns matching catalog entries with their available items
// populated, so the caller can render stock counts directly.
func (s *Service) ListProducts(ctx context.Context, in ListProductsInput) ([]*entity.Product, error) {
	if in.Limit <= 0 {
		in.Limit = 50
	}
	products, err := s.productRepo.GetByBranchID(ctx, in.BranchID, in.Kind, in.Search, in.Limit, in.Offset)
	if err != nil {
		return nil, err
	}
	for _, p := range products {
		items, _ := s.itemRepo.GetByProductID(ctx, p.ID, []entity.ProductStatus{entity.ProductStatusAvailable})
		p.Items = items
	}
	return products, nil
}

// ── ProductItem CRUD ──────────────────────────────────────────────────────────

// CreateItemInput represents a single physical piece being added to stock.
// Weight 0 with kind=bar will fall back to BarSizeBaht × 15.244.
type CreateItemInput struct {
	ProductID    primitive.ObjectID
	Barcode      string
	SerialNumber string
	WeightGrams  float64
	LaborCost    *float64 // nil → fall back to product.DefaultItemLaborCost()
	Cost         float64
	Note         string
	By           primitive.ObjectID
}

// CreateItem adds one physical item to stock and writes a stock log.
func (s *Service) CreateItem(ctx context.Context, in CreateItemInput) (*entity.ProductItem, error) {
	if in.Barcode == "" {
		return nil, errors.New("barcode is required")
	}
	if existing, _ := s.itemRepo.GetByBarcode(ctx, in.Barcode); existing != nil {
		return nil, errors.New("barcode already in use")
	}
	product, err := s.productRepo.GetByID(ctx, in.ProductID)
	if err != nil || product == nil {
		return nil, errors.New("product not found")
	}
	weight := in.WeightGrams
	if weight <= 0 {
		weight = product.DefaultItemWeightGrams()
		if weight <= 0 {
			return nil, errors.New("weight is required")
		}
	}
	labor := product.DefaultItemLaborCost()
	if in.LaborCost != nil {
		labor = *in.LaborCost
	}

	item := entity.NewProductItem(in.ProductID, product.BranchID, in.Barcode, weight, labor, in.Cost)
	item.SerialNumber = in.SerialNumber
	item.Note = in.Note

	if err := s.itemRepo.Create(ctx, item); err != nil {
		return nil, err
	}
	s.logStock(ctx, product.BranchID, in.ProductID, &item.ID, in.By, entity.StockActionAdd, weight)
	return item, nil
}

// BulkCreateItemsInput is the bar-friendly path: "I just received 5 bars of 1 baht each".
// For ornaments, prefer per-item creation since each piece is unique.
type BulkCreateItemsInput struct {
	ProductID   primitive.ObjectID
	Count       int
	WeightGrams float64 // 0 → falls back per-product (bar default size)
	LaborCost   *float64
	Cost        float64
	BarcodeSeed string // used as prefix; system appends -1, -2, ...
	By          primitive.ObjectID
}

// BulkCreateItems inserts `Count` items. Barcodes are derived from BarcodeSeed
// or a random prefix if BarcodeSeed is empty.
func (s *Service) BulkCreateItems(ctx context.Context, in BulkCreateItemsInput) ([]*entity.ProductItem, error) {
	if in.Count <= 0 {
		return nil, errors.New("count must be positive")
	}
	product, err := s.productRepo.GetByID(ctx, in.ProductID)
	if err != nil || product == nil {
		return nil, errors.New("product not found")
	}
	weight := in.WeightGrams
	if weight <= 0 {
		weight = product.DefaultItemWeightGrams()
		if weight <= 0 {
			return nil, errors.New("weight is required for ornament bulk add")
		}
	}
	labor := product.DefaultItemLaborCost()
	if in.LaborCost != nil {
		labor = *in.LaborCost
	}
	seed := in.BarcodeSeed
	if seed == "" {
		seed = product.SKU + "-" + time.Now().Format("060102")
	}

	out := make([]*entity.ProductItem, 0, in.Count)
	for i := 1; i <= in.Count; i++ {
		barcode := fmt.Sprintf("%s-%03d", seed, i)
		// Skip if already used (idempotent re-runs).
		if existing, _ := s.itemRepo.GetByBarcode(ctx, barcode); existing != nil {
			continue
		}
		item := entity.NewProductItem(in.ProductID, product.BranchID, barcode, weight, labor, in.Cost)
		if err := s.itemRepo.Create(ctx, item); err != nil {
			return nil, err
		}
		s.logStock(ctx, product.BranchID, in.ProductID, &item.ID, in.By, entity.StockActionAdd, weight)
		out = append(out, item)
	}
	return out, nil
}

// UpdateItemInput updates a single physical piece. Status is not exposed here —
// changes go through sale/cancel/transfer flows.
type UpdateItemInput struct {
	WeightGrams  *float64
	LaborCost    *float64
	Cost         *float64
	SerialNumber *string
	Note         *string
}

func (s *Service) UpdateItem(ctx context.Context, itemID primitive.ObjectID, in UpdateItemInput) (*entity.ProductItem, error) {
	item, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil || item == nil {
		return nil, errors.New("item not found")
	}
	if item.Status != entity.ProductStatusAvailable {
		return nil, errors.New("only available items can be edited")
	}
	if in.WeightGrams != nil && *in.WeightGrams > 0 {
		item.WeightGrams = *in.WeightGrams
	}
	if in.LaborCost != nil {
		item.LaborCost = *in.LaborCost
	}
	if in.Cost != nil {
		item.Cost = *in.Cost
	}
	if in.SerialNumber != nil {
		item.SerialNumber = *in.SerialNumber
	}
	if in.Note != nil {
		item.Note = *in.Note
	}
	if err := s.itemRepo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) DeleteItem(ctx context.Context, itemID primitive.ObjectID, by primitive.ObjectID) error {
	item, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil || item == nil {
		return errors.New("item not found")
	}
	if item.Status != entity.ProductStatusAvailable {
		return errors.New("only available items can be deleted")
	}
	if err := s.itemRepo.Delete(ctx, itemID); err != nil {
		return err
	}
	s.logStock(ctx, item.BranchID, item.ProductID, &itemID, by, entity.StockActionRemove, item.WeightGrams)
	return nil
}

// ListItems returns physical items for a product (default: only available).
func (s *Service) ListItems(ctx context.Context, productID primitive.ObjectID, status []entity.ProductStatus) ([]*entity.ProductItem, error) {
	if status == nil {
		status = []entity.ProductStatus{entity.ProductStatusAvailable}
	}
	return s.itemRepo.GetByProductID(ctx, productID, status)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func validateCreate(in *CreateProductInput) error {
	if in.SKU == "" {
		return errors.New("SKU is required")
	}
	if in.Name == "" {
		return errors.New("name is required")
	}
	if in.GoldType == "" {
		return errors.New("gold_type is required")
	}
	switch in.Kind {
	case entity.KindOrnament:
		if in.DefaultLaborCost < 0 {
			return errors.New("default_labor_cost must be non-negative")
		}
	case entity.KindBar:
		if in.BarSizeBaht == nil || *in.BarSizeBaht <= 0 {
			return errors.New("bar_size_baht must be positive")
		}
	default:
		return errors.New("kind must be ornament or bar")
	}
	return nil
}

func (s *Service) logStock(
	ctx context.Context, branchID, productID primitive.ObjectID,
	itemID *primitive.ObjectID, by primitive.ObjectID,
	action entity.StockAction, weight float64,
) {
	log := entity.NewStockLog(branchID, productID, by, action, weight)
	log.ProductItemID = itemID
	_ = s.stockLogRepo.Create(ctx, log)
}
