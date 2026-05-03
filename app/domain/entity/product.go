package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProductStatus represents the status of a single physical gold item.
type ProductStatus string

const (
	ProductStatusAvailable ProductStatus = "available"
	ProductStatusSold      ProductStatus = "sold"
	ProductStatusReserved  ProductStatus = "reserved"
	ProductStatusPawned    ProductStatus = "pawned"
)

// ProductKind discriminates how a product is priced and described.
//
//	ornament — ทองรูปพรรณ; per-piece labor (ค่ากำเหน็จ); free-text Design.
//	bar      — ทองคำแท่ง; no labor; weight derived from BarSizeBaht (× 15.244 g/baht).
type ProductKind string

const (
	KindOrnament ProductKind = "ornament"
	KindBar      ProductKind = "bar"
)

const (
	// 1 บาททอง = 15.244 g (Thai market standard, applies to both 96.5% and 99.99%).
	BahtPerGram = 15.244
	// Legacy aliases retained for sale.go imports while we migrate; both equal 15.244.
	BahtPerGramOrnament = BahtPerGram
	BahtPerGramBar      = BahtPerGram
)

// IsBarGold returns true when the product is sold as a gold bar.
func (p *Product) IsBarGold() bool {
	return p.Kind == KindBar
}

// Product is the catalog entry. It holds no balance and no master-level price;
// every sellable unit lives in a ProductItem.
//
//   - Kind determines required/optional fields:
//   - ornament → Design (string), DefaultLaborCost
//   - bar      → BarSizeBaht (float, free input)
//   - GoldType is informational ("96.5%" / "99.99%").
type Product struct {
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	BranchID         primitive.ObjectID `json:"branch_id" bson:"branch_id"`
	SKU              string             `json:"sku" bson:"sku"`
	Code             string             `json:"code,omitempty" bson:"code,omitempty"`
	Kind             ProductKind        `json:"kind" bson:"kind"`
	GoldType         string             `json:"gold_type" bson:"gold_type"`
	Name             string             `json:"name" bson:"name"`
	Description      string             `json:"description,omitempty" bson:"description,omitempty"`
	Design           string             `json:"design,omitempty" bson:"design,omitempty"`               // ornament-only
	BarSizeBaht      *float64           `json:"bar_size_baht,omitempty" bson:"bar_size_baht,omitempty"` // bar-only
	DefaultLaborCost float64            `json:"default_labor_cost" bson:"default_labor_cost"`           // ornament-only
	Images           []string           `json:"images" bson:"images"`
	IsActive         bool               `json:"is_active" bson:"is_active"`
	Items            []*ProductItem     `json:"items,omitempty" bson:"-"` // populated at read time, not persisted on Product

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// NewOrnamentProduct creates an ornament catalog entry.
func NewOrnamentProduct(branchID primitive.ObjectID, sku, name, design, goldType string, defaultLabor float64) *Product {
	now := time.Now()
	return &Product{
		BranchID:         branchID,
		SKU:              sku,
		Kind:             KindOrnament,
		GoldType:         goldType,
		Name:             name,
		Design:           design,
		DefaultLaborCost: defaultLabor,
		Images:           []string{},
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// NewBarProduct creates a gold-bar catalog entry pinned to a baht size.
// The pointer to BarSizeBaht is set so omitempty correctly distinguishes "missing".
func NewBarProduct(branchID primitive.ObjectID, sku, name, goldType string, barSizeBaht float64) *Product {
	now := time.Now()
	size := barSizeBaht
	return &Product{
		BranchID:    branchID,
		SKU:         sku,
		Kind:        KindBar,
		GoldType:    goldType,
		Name:        name,
		BarSizeBaht: &size,
		Images:      []string{},
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// DefaultItemWeightGrams returns the canonical weight for a freshly-created item:
//   - bar: derived from BarSizeBaht × 15.244 (the operator may override per item)
//   - ornament: 0 (operator must enter actual weight at item creation)
func (p *Product) DefaultItemWeightGrams() float64 {
	if p.Kind == KindBar && p.BarSizeBaht != nil {
		return *p.BarSizeBaht * BahtPerGram
	}
	return 0
}

// DefaultItemLaborCost returns the labor seed for a new item:
//   - ornament: DefaultLaborCost from the master
//   - bar:      always 0
func (p *Product) DefaultItemLaborCost() float64 {
	if p.Kind == KindBar {
		return 0
	}
	return p.DefaultLaborCost
}
