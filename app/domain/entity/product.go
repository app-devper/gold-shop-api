package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProductStatus represents the status of a product
type ProductStatus string

const (
	ProductStatusAvailable ProductStatus = "available"
	ProductStatusSold      ProductStatus = "sold"
	ProductStatusReserved  ProductStatus = "reserved"
	ProductStatusPawned    ProductStatus = "pawned"
)

// StockType represents how a product's stock is counted
type StockType string

const (
	StockTypePiece  StockType = "piece"
	StockTypeWeight StockType = "weight"
)

const (
	// BahtPerGramOrnament is the baht-to-gram conversion factor for gold ornaments (96.5%)
	BahtPerGramOrnament = 15.16
	// BahtPerGramBar is the baht-to-gram conversion factor for gold bars (99.99%)
	BahtPerGramBar = 15.244
)

// IsBarGold returns true if the product's gold type is bar gold (99.99%)
func (p *Product) IsBarGold() bool {
	return p.GoldType == "99.99%" || p.GoldType == "99.99"
}

// DefaultStockType returns the effective StockType, defaulting to StockTypeWeight for legacy products
func (p *Product) DefaultStockType() StockType {
	if p.StockType == "" {
		return StockTypeWeight
	}
	return p.StockType
}

// ProductCategory represents a product category
type ProductCategory struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Code        string             `json:"code" bson:"code"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	IsActive    bool               `json:"is_active" bson:"is_active"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// Product represents a gold product
type Product struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	BranchID     primitive.ObjectID `json:"branch_id" bson:"branch_id"`
	CategoryID   primitive.ObjectID `json:"category_id" bson:"category_id"`
	SKU          string             `json:"sku" bson:"sku"`
	Barcode      string             `json:"barcode,omitempty" bson:"barcode,omitempty"`
	Name         string             `json:"name" bson:"name"`
	Description  string             `json:"description" bson:"description"`
	StockType    StockType          `json:"stock_type" bson:"stock_type"`
	GoldType     string             `json:"gold_type" bson:"gold_type"`     // 96.5%, 99.99%
	Weight       float64            `json:"weight" bson:"weight"`           // grams
	WeightUnit   string             `json:"weight_unit" bson:"weight_unit"` // baht, gram
	LaborCost    float64            `json:"labor_cost" bson:"labor_cost"`   // ค่ากำเหน็จ
	Price        float64            `json:"price" bson:"price"`
	Cost         float64            `json:"cost" bson:"cost"`
	Status       ProductStatus      `json:"status" bson:"status"`
	Images       []string           `json:"images" bson:"images"`
	ReorderPoint int                `json:"reorder_point" bson:"reorder_point"`
	Items        []*ProductItem     `json:"items,omitempty" bson:"-"` // Not stored in product collection
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

// NewProduct creates a new Product entity
func NewProduct(branchID, categoryID primitive.ObjectID, sku, name, goldType string, stockType StockType, weight, laborCost float64) *Product {
	now := time.Now()
	return &Product{
		BranchID:   branchID,
		CategoryID: categoryID,
		SKU:        sku,
		Name:       name,
		StockType:  stockType,
		GoldType:   goldType,
		Weight:     weight,
		WeightUnit: "gram",
		LaborCost:  laborCost,
		Status:     ProductStatusAvailable,
		Images:     []string{},
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// GetPrice returns the product price
func (p *Product) GetPrice() float64 {
	return p.Price
}

// CalculateTotalPrice calculates total price including labor cost
func (p *Product) CalculateTotalPrice() float64 {
	return p.Price + p.LaborCost
}
