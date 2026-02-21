package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TransferStatus represents inventory transfer status
type TransferStatus string

const (
	TransferStatusPending   TransferStatus = "pending"
	TransferStatusInTransit TransferStatus = "in_transit"
	TransferStatusReceived  TransferStatus = "received"
	TransferStatusCancelled TransferStatus = "cancelled"
)

// TransferItem represents an item in a transfer
type TransferItem struct {
	ProductID primitive.ObjectID `json:"product_id" bson:"product_id"`
	Quantity  int                `json:"quantity" bson:"quantity"`
}

// InventoryTransfer represents a stock transfer between branches
type InventoryTransfer struct {
	ID             primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	TransferNumber string              `json:"transfer_number" bson:"transfer_number"`
	FromBranchID   primitive.ObjectID  `json:"from_branch_id" bson:"from_branch_id"`
	ToBranchID     primitive.ObjectID  `json:"to_branch_id" bson:"to_branch_id"`
	Items          []TransferItem      `json:"items" bson:"items"`
	Status         TransferStatus      `json:"status" bson:"status"`
	RequestedBy    primitive.ObjectID  `json:"requested_by" bson:"requested_by"`
	ApprovedBy     *primitive.ObjectID `json:"approved_by,omitempty" bson:"approved_by,omitempty"`
	ReceivedBy     *primitive.ObjectID `json:"received_by,omitempty" bson:"received_by,omitempty"`
	Notes          string              `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedAt      time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at" bson:"updated_at"`
}

// NewInventoryTransfer creates a new InventoryTransfer entity
func NewInventoryTransfer(transferNumber string, fromBranchID, toBranchID, requestedBy primitive.ObjectID) *InventoryTransfer {
	now := time.Now()
	return &InventoryTransfer{
		TransferNumber: transferNumber,
		FromBranchID:   fromBranchID,
		ToBranchID:     toBranchID,
		Items:          []TransferItem{},
		Status:         TransferStatusPending,
		RequestedBy:    requestedBy,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// AddItem adds an item to the transfer
func (t *InventoryTransfer) AddItem(productID primitive.ObjectID, quantity int) {
	t.Items = append(t.Items, TransferItem{
		ProductID: productID,
		Quantity:  quantity,
	})
	t.UpdatedAt = time.Now()
}

// Approve approves the transfer
func (t *InventoryTransfer) Approve(approvedBy primitive.ObjectID) {
	t.ApprovedBy = &approvedBy
	t.Status = TransferStatusInTransit
	t.UpdatedAt = time.Now()
}

// Receive marks the transfer as received
func (t *InventoryTransfer) Receive(receivedBy primitive.ObjectID) {
	t.ReceivedBy = &receivedBy
	t.Status = TransferStatusReceived
	t.UpdatedAt = time.Now()
}

// Cancel cancels the transfer
func (t *InventoryTransfer) Cancel() {
	t.Status = TransferStatusCancelled
	t.UpdatedAt = time.Now()
}
