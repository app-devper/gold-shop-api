package entity

import (
	"time"

	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SaleType represents different types of sales
type SaleType string

const (
	SaleTypeSell     SaleType = "sell"
	SaleTypeBuyOld   SaleType = "buy_old"
	SaleTypeExchange SaleType = "exchange"
)

// SaleStatus represents sale transaction status
type SaleStatus string

const (
	SaleStatusCompleted SaleStatus = "completed"
	SaleStatusPending   SaleStatus = "pending"
	SaleStatusCancelled SaleStatus = "cancelled"
)

// DiscountType represents discount calculation type
type DiscountType string

const (
	DiscountTypeAmount  DiscountType = "amount"
	DiscountTypePercent DiscountType = "percent"
)

// PaymentMethod represents payment methods
type PaymentMethod string

const (
	PaymentMethodCash       PaymentMethod = "cash"
	PaymentMethodCreditCard PaymentMethod = "credit_card"
	PaymentMethodTransfer   PaymentMethod = "transfer"
	PaymentMethodVoucher    PaymentMethod = "voucher"
)

// SaleItem represents an item in a sale
type SaleItem struct {
	ProductID     primitive.ObjectID  `json:"product_id" bson:"product_id"`
	ProductItemID *primitive.ObjectID `json:"product_item_id,omitempty" bson:"product_item_id,omitempty"`
	ProductName   string              `json:"product_name" bson:"product_name"`
	GoldType      string              `json:"gold_type" bson:"gold_type"`
	Weight        float64             `json:"weight" bson:"weight"`
	PriceLevel    string              `json:"price_level" bson:"price_level"`
	UnitPrice     float64             `json:"unit_price" bson:"unit_price"`
	LaborCost     float64             `json:"labor_cost" bson:"labor_cost"`
	Discount      float64             `json:"discount" bson:"discount"`
	DiscountType  DiscountType        `json:"discount_type" bson:"discount_type"`
	Cost          float64             `json:"cost" bson:"cost"`
	Total         float64             `json:"total" bson:"total"`
}

// OldGoldItem represents old gold being traded in
type OldGoldItem struct {
	Description  string  `json:"description" bson:"description"`
	GoldType     string  `json:"gold_type" bson:"gold_type"`
	Weight       float64 `json:"weight" bson:"weight"`
	PricePerUnit float64 `json:"price_per_unit" bson:"price_per_unit"`
	Total        float64 `json:"total" bson:"total"`
}

// Payment represents a payment in a sale
type Payment struct {
	Method    PaymentMethod `json:"method" bson:"method"`
	Amount    float64       `json:"amount" bson:"amount"`
	Reference string        `json:"reference,omitempty" bson:"reference,omitempty"`
}

// Sale represents a sales transaction
type Sale struct {
	ID           primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	BranchID     primitive.ObjectID  `json:"branch_id" bson:"branch_id"`
	SaleNumber   string              `json:"sale_number" bson:"sale_number"`
	CustomerID   *primitive.ObjectID `json:"customer_id,omitempty" bson:"customer_id,omitempty"`
	UserID       primitive.ObjectID  `json:"user_id" bson:"user_id"`
	SaleType     SaleType            `json:"sale_type" bson:"sale_type"`
	Items        []SaleItem          `json:"items" bson:"items"`
	OldGoldItems []OldGoldItem       `json:"old_gold_items,omitempty" bson:"old_gold_items,omitempty"`
	Subtotal     float64             `json:"subtotal" bson:"subtotal"`
	Discount     float64             `json:"discount" bson:"discount"`
	DiscountType DiscountType        `json:"discount_type" bson:"discount_type"`
	OldGoldValue float64             `json:"old_gold_value" bson:"old_gold_value"`
	NetTotal     float64             `json:"net_total" bson:"net_total"`
	Payments     []Payment           `json:"payments" bson:"payments"`
	PointsEarned int                 `json:"points_earned" bson:"points_earned"`
	PointsUsed   int                 `json:"points_used" bson:"points_used"`
	Status       SaleStatus          `json:"status" bson:"status"`
	Notes        string              `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedAt    time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at" bson:"updated_at"`
}

// NewSale creates a new Sale entity
func NewSale(branchID, userID primitive.ObjectID, saleNumber string, saleType SaleType) *Sale {
	now := time.Now()
	return &Sale{
		BranchID:   branchID,
		SaleNumber: saleNumber,
		UserID:     userID,
		SaleType:   saleType,
		Items:      []SaleItem{},
		Payments:   []Payment{},
		Status:     SaleStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// AddItem adds an item to the sale
func (s *Sale) AddItem(item SaleItem) {
	s.Items = append(s.Items, item)
	s.CalculateTotals()
}

// AddOldGoldItem adds an old gold item for exchange
func (s *Sale) AddOldGoldItem(item OldGoldItem) {
	s.OldGoldItems = append(s.OldGoldItems, item)
	s.CalculateTotals()
}

// AddPayment adds a payment to the sale
func (s *Sale) AddPayment(payment Payment) {
	s.Payments = append(s.Payments, payment)
}

// CalculateTotals recalculates all totals
func (s *Sale) CalculateTotals() {
	subtotal := 0.0
	for _, item := range s.Items {
		subtotal += item.Total
	}
	s.Subtotal = utils.RoundBaht(subtotal)

	oldGold := 0.0
	for _, item := range s.OldGoldItems {
		oldGold += item.Total
	}
	s.OldGoldValue = utils.RoundBaht(oldGold)

	discountAmount := s.Discount
	if s.DiscountType == DiscountTypePercent {
		discountAmount = s.Subtotal * (s.Discount / 100)
	}

	net := s.Subtotal - discountAmount - s.OldGoldValue
	if net < 0 {
		net = 0
	}
	s.NetTotal = utils.RoundBaht(net)
}

// GetTotalPayments returns total amount paid
func (s *Sale) GetTotalPayments() float64 {
	total := 0.0
	for _, p := range s.Payments {
		total += p.Amount
	}
	return total
}

// IsFullyPaid checks if sale is fully paid
func (s *Sale) IsFullyPaid() bool {
	return s.GetTotalPayments() >= s.NetTotal
}

// Complete marks the sale as completed
func (s *Sale) Complete() {
	s.Status = SaleStatusCompleted
	s.UpdatedAt = time.Now()
}

// Cancel cancels the sale
func (s *Sale) Cancel() {
	s.Status = SaleStatusCancelled
	s.UpdatedAt = time.Now()
}
