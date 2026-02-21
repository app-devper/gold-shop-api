package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ExpenseStatus represents expense approval status
type ExpenseStatus string

const (
	ExpenseStatusPending  ExpenseStatus = "pending"
	ExpenseStatusApproved ExpenseStatus = "approved"
	ExpenseStatusRejected ExpenseStatus = "rejected"
)

// ExpenseCategory represents an expense category
type ExpenseCategory struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Code      string             `json:"code" bson:"code"`
	Name      string             `json:"name" bson:"name"`
	IsActive  bool               `json:"is_active" bson:"is_active"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// Expense represents an expense record
type Expense struct {
	ID            primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	BranchID      primitive.ObjectID  `json:"branch_id" bson:"branch_id"`
	CategoryID    primitive.ObjectID  `json:"category_id" bson:"category_id"`
	ExpenseNumber string              `json:"expense_number" bson:"expense_number"`
	Description   string              `json:"description" bson:"description"`
	Amount        float64             `json:"amount" bson:"amount"`
	ExpenseDate   time.Time           `json:"expense_date" bson:"expense_date"`
	ReceiptNumber string              `json:"receipt_number,omitempty" bson:"receipt_number,omitempty"`
	Attachments   []string            `json:"attachments" bson:"attachments"`
	CreatedBy     primitive.ObjectID  `json:"created_by" bson:"created_by"`
	ApprovedBy    *primitive.ObjectID `json:"approved_by,omitempty" bson:"approved_by,omitempty"`
	Status        ExpenseStatus       `json:"status" bson:"status"`
	Notes         string              `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedAt     time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at" bson:"updated_at"`
}

// NewExpense creates a new Expense entity
func NewExpense(branchID, categoryID, createdBy primitive.ObjectID, expenseNumber, description string, amount float64, expenseDate time.Time) *Expense {
	now := time.Now()
	return &Expense{
		BranchID:      branchID,
		CategoryID:    categoryID,
		ExpenseNumber: expenseNumber,
		Description:   description,
		Amount:        amount,
		ExpenseDate:   expenseDate,
		Attachments:   []string{},
		CreatedBy:     createdBy,
		Status:        ExpenseStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// Approve approves the expense
func (e *Expense) Approve(approvedBy primitive.ObjectID) {
	e.ApprovedBy = &approvedBy
	e.Status = ExpenseStatusApproved
	e.UpdatedAt = time.Now()
}

// Reject rejects the expense
func (e *Expense) Reject(approvedBy primitive.ObjectID, notes string) {
	e.ApprovedBy = &approvedBy
	e.Status = ExpenseStatusRejected
	e.Notes = notes
	e.UpdatedAt = time.Now()
}

// AddAttachment adds an attachment to the expense
func (e *Expense) AddAttachment(path string) {
	e.Attachments = append(e.Attachments, path)
	e.UpdatedAt = time.Now()
}
