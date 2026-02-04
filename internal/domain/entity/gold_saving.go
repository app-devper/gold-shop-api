package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GoldSavingType represents the type of gold saving account
type GoldSavingType string

const (
	GoldSavingByMoney  GoldSavingType = "by_money"
	GoldSavingByWeight GoldSavingType = "by_weight"
)

// GoldSavingStatus represents the status of a gold saving account
type GoldSavingStatus string

const (
	GoldSavingStatusActive GoldSavingStatus = "active"
	GoldSavingStatusClosed GoldSavingStatus = "closed"
)

// TransactionType represents gold saving transaction type
type TransactionType string

const (
	TransactionDeposit    TransactionType = "deposit"
	TransactionWithdrawal TransactionType = "withdrawal"
)

// GoldSavingTransaction represents a transaction in gold saving
type GoldSavingTransaction struct {
	Date         time.Time          `json:"date" bson:"date"`
	Type         TransactionType    `json:"type" bson:"type"`
	Amount       float64            `json:"amount" bson:"amount"`           // money or weight
	GoldPrice    float64            `json:"gold_price" bson:"gold_price"`   // price at transaction time
	GoldWeight   float64            `json:"gold_weight" bson:"gold_weight"` // calculated gold weight
	BalanceAfter float64            `json:"balance_after" bson:"balance_after"`
	ProcessedBy  primitive.ObjectID `json:"processed_by" bson:"processed_by"`
}

// GoldSaving represents a gold saving account (ออมทอง)
type GoldSaving struct {
	ID            primitive.ObjectID      `json:"id" bson:"_id,omitempty"`
	BranchID      primitive.ObjectID      `json:"branch_id" bson:"branch_id"`
	AccountNumber string                  `json:"account_number" bson:"account_number"`
	CustomerID    primitive.ObjectID      `json:"customer_id" bson:"customer_id"`
	SavingType    GoldSavingType          `json:"saving_type" bson:"saving_type"`
	MinDeposit    float64                 `json:"min_deposit" bson:"min_deposit"`
	MinWithdrawal float64                 `json:"min_withdrawal" bson:"min_withdrawal"`
	GoldBalance   float64                 `json:"gold_balance" bson:"gold_balance"` // grams
	CashBalance   float64                 `json:"cash_balance" bson:"cash_balance"`
	Transactions  []GoldSavingTransaction `json:"transactions" bson:"transactions"`
	Status        GoldSavingStatus        `json:"status" bson:"status"`
	OpenedDate    time.Time               `json:"opened_date" bson:"opened_date"`
	ClosedDate    *time.Time              `json:"closed_date,omitempty" bson:"closed_date,omitempty"`
	CreatedAt     time.Time               `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at" bson:"updated_at"`
}

// NewGoldSaving creates a new GoldSaving account
func NewGoldSaving(branchID, customerID primitive.ObjectID, accountNumber string, savingType GoldSavingType, minDeposit, minWithdrawal float64) *GoldSaving {
	now := time.Now()
	return &GoldSaving{
		BranchID:      branchID,
		AccountNumber: accountNumber,
		CustomerID:    customerID,
		SavingType:    savingType,
		MinDeposit:    minDeposit,
		MinWithdrawal: minWithdrawal,
		GoldBalance:   0,
		CashBalance:   0,
		Transactions:  []GoldSavingTransaction{},
		Status:        GoldSavingStatusActive,
		OpenedDate:    now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// Deposit adds money/gold to the saving account
func (gs *GoldSaving) Deposit(amount, goldPrice float64, processedBy primitive.ObjectID) error {
	var goldWeight float64

	if gs.SavingType == GoldSavingByMoney {
		// Calculate gold weight based on current price
		goldWeight = amount / goldPrice
		gs.GoldBalance += goldWeight
		gs.CashBalance += amount
	} else {
		// Direct gold weight deposit
		goldWeight = amount
		gs.GoldBalance += goldWeight
	}

	transaction := GoldSavingTransaction{
		Date:         time.Now(),
		Type:         TransactionDeposit,
		Amount:       amount,
		GoldPrice:    goldPrice,
		GoldWeight:   goldWeight,
		BalanceAfter: gs.GoldBalance,
		ProcessedBy:  processedBy,
	}

	gs.Transactions = append(gs.Transactions, transaction)
	gs.UpdatedAt = time.Now()

	return nil
}

// Withdraw removes gold/money from the saving account
func (gs *GoldSaving) Withdraw(amount float64, asCash bool, goldPrice float64, processedBy primitive.ObjectID) error {
	var goldWeight float64

	if asCash {
		// Withdraw as cash - calculate gold weight to deduct
		goldWeight = amount / goldPrice
	} else {
		// Withdraw as physical gold
		goldWeight = amount
	}

	if goldWeight > gs.GoldBalance {
		return ErrInsufficientBalance
	}

	gs.GoldBalance -= goldWeight

	transaction := GoldSavingTransaction{
		Date:         time.Now(),
		Type:         TransactionWithdrawal,
		Amount:       amount,
		GoldPrice:    goldPrice,
		GoldWeight:   goldWeight,
		BalanceAfter: gs.GoldBalance,
		ProcessedBy:  processedBy,
	}

	gs.Transactions = append(gs.Transactions, transaction)
	gs.UpdatedAt = time.Now()

	return nil
}

// GetCurrentValue returns current value of gold savings
func (gs *GoldSaving) GetCurrentValue(currentGoldPrice float64) float64 {
	return gs.GoldBalance * currentGoldPrice
}

// Close closes the gold saving account
func (gs *GoldSaving) Close() {
	now := time.Now()
	gs.Status = GoldSavingStatusClosed
	gs.ClosedDate = &now
	gs.UpdatedAt = now
}
