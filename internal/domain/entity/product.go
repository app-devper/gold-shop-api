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

// ProductPrices contains tiered pricing for products
type ProductPrices struct {
	LevelA float64 `json:"level_a" bson:"level_a"`
	LevelB float64 `json:"level_b" bson:"level_b"`
	LevelC float64 `json:"level_c" bson:"level_c"`
	LevelD float64 `json:"level_d" bson:"level_d"`
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
	GoldType     string             `json:"gold_type" bson:"gold_type"`     // 96.5%, 99.99%
	Weight       float64            `json:"weight" bson:"weight"`           // grams
	WeightUnit   string             `json:"weight_unit" bson:"weight_unit"` // baht, gram
	LaborCost    float64            `json:"labor_cost" bson:"labor_cost"`   // ค่ากำเหน็จ
	Prices       ProductPrices      `json:"prices" bson:"prices"`
	Cost         float64            `json:"cost" bson:"cost"`
	Status       ProductStatus      `json:"status" bson:"status"`
	Images       []string           `json:"images" bson:"images"`
	ReorderPoint int                `json:"reorder_point" bson:"reorder_point"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

// NewProduct creates a new Product entity
func NewProduct(branchID, categoryID primitive.ObjectID, sku, name, goldType string, weight, laborCost float64) *Product {
	now := time.Now()
	return &Product{
		BranchID:   branchID,
		CategoryID: categoryID,
		SKU:        sku,
		Name:       name,
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

// GetPriceByLevel returns price for the specified level
func (p *Product) GetPriceByLevel(level string) float64 {
	switch level {
	case "A":
		return p.Prices.LevelA
	case "B":
		return p.Prices.LevelB
	case "C":
		return p.Prices.LevelC
	case "D":
		return p.Prices.LevelD
	default:
		return p.Prices.LevelA
	}
}

// CalculateTotalPrice calculates total price including labor cost
func (p *Product) CalculateTotalPrice(level string) float64 {
	return p.GetPriceByLevel(level) + p.LaborCost
}
