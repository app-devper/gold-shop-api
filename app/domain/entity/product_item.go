package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProductItem is a physical sellable instance of a Product. Every gold piece in
// stock is one of these — for both ornaments and bars — keyed by Barcode.
type ProductItem struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ProductID    primitive.ObjectID `json:"product_id" bson:"product_id"`
	BranchID     primitive.ObjectID `json:"branch_id" bson:"branch_id"`
	Barcode      string             `json:"barcode" bson:"barcode"`
	SerialNumber string             `json:"serial_number,omitempty" bson:"serial_number,omitempty"`
	WeightGrams  float64            `json:"weight_grams" bson:"weight_grams"`
	LaborCost    float64            `json:"labor_cost" bson:"labor_cost"`
	Cost         float64            `json:"cost" bson:"cost"`
	Status       ProductStatus      `json:"status" bson:"status"`
	ReceivedDate time.Time          `json:"received_date" bson:"received_date"`
	Note         string             `json:"note,omitempty" bson:"note,omitempty"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

// NewProductItem builds a fresh available item.
func NewProductItem(productID, branchID primitive.ObjectID, barcode string, weightGrams, laborCost, cost float64) *ProductItem {
	now := time.Now()
	return &ProductItem{
		ProductID:    productID,
		BranchID:     branchID,
		Barcode:      barcode,
		WeightGrams:  weightGrams,
		LaborCost:    laborCost,
		Cost:         cost,
		Status:       ProductStatusAvailable,
		ReceivedDate: now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// Backwards-compat aliases retained while sale.go is migrated.
func (pi *ProductItem) Weight() float64 { return pi.WeightGrams }
