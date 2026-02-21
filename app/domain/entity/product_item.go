package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProductItem represents an individual physical item of a Product (e.g., a specific gold necklace)
type ProductItem struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ProductID primitive.ObjectID `json:"product_id" bson:"product_id"`
	BranchID  primitive.ObjectID `json:"branch_id" bson:"branch_id"`
	Barcode   string             `json:"barcode" bson:"barcode"`
	Weight    float64            `json:"weight" bson:"weight"`
	LaborCost float64            `json:"labor_cost" bson:"labor_cost"`
	Status    ProductStatus      `json:"status" bson:"status"`
	Notes     string             `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// NewProductItem creates a new ProductItem entity
func NewProductItem(productID, branchID primitive.ObjectID, barcode string, weight, laborCost float64) *ProductItem {
	now := time.Now()
	return &ProductItem{
		ProductID: productID,
		BranchID:  branchID,
		Barcode:   barcode,
		Weight:    weight,
		LaborCost: laborCost,
		Status:    ProductStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
