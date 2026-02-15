package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StockAction represents the type of stock movement
type StockAction string

const (
	StockActionAdd      StockAction = "add"
	StockActionRemove   StockAction = "remove"
	StockActionSale     StockAction = "sale"
	StockActionCancel   StockAction = "cancel"
	StockActionTransfer StockAction = "transfer"
	StockActionAdjust   StockAction = "adjust"
)

// StockLog represents an audit log for stock changes
type StockLog struct {
	ID            primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	BranchID      primitive.ObjectID  `json:"branch_id" bson:"branch_id"`
	ProductID     primitive.ObjectID  `json:"product_id" bson:"product_id"`
	ProductItemID *primitive.ObjectID `json:"product_item_id,omitempty" bson:"product_item_id,omitempty"`
	Action        StockAction         `json:"action" bson:"action"`
	Quantity      float64             `json:"quantity" bson:"quantity"` // Pieces or Weight
	ReferenceID   string              `json:"reference_id,omitempty" bson:"reference_id,omitempty"`
	UserID        primitive.ObjectID  `json:"user_id" bson:"user_id"`
	Notes         string              `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedAt     time.Time           `json:"created_at" bson:"created_at"`
}

// NewStockLog creates a new StockLog entity
func NewStockLog(branchID, productID, userID primitive.ObjectID, action StockAction, quantity float64) *StockLog {
	return &StockLog{
		BranchID:  branchID,
		ProductID: productID,
		UserID:    userID,
		Action:    action,
		Quantity:  quantity,
		CreatedAt: time.Now(),
	}
}
