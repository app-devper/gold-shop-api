package entity

import (
	"time"

	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PawnStatus represents the status of a pawn
type PawnStatus string

const (
	PawnStatusActive    PawnStatus = "active"
	PawnStatusRedeemed  PawnStatus = "redeemed"
	PawnStatusForfeited PawnStatus = "forfeited"
	PawnStatusExtended  PawnStatus = "extended"
)

// PawnItem represents an item in a pawn transaction
type PawnItem struct {
	Description    string   `json:"description" bson:"description"`
	GoldType       string   `json:"gold_type" bson:"gold_type"`
	Weight         float64  `json:"weight" bson:"weight"`
	AppraisedValue float64  `json:"appraised_value" bson:"appraised_value"`
	Images         []string `json:"images" bson:"images"`
}

// InterestPayment represents an interest payment record
type InterestPayment struct {
	PaymentDate time.Time          `json:"payment_date" bson:"payment_date"`
	Amount      float64            `json:"amount" bson:"amount"`
	PeriodFrom  time.Time          `json:"period_from" bson:"period_from"`
	PeriodTo    time.Time          `json:"period_to" bson:"period_to"`
	ReceivedBy  primitive.ObjectID `json:"received_by" bson:"received_by"`
}

// Redemption represents pawn redemption details
type Redemption struct {
	Date       time.Time          `json:"date" bson:"date"`
	Principal  float64            `json:"principal" bson:"principal"`
	Interest   float64            `json:"interest" bson:"interest"`
	Discount   float64            `json:"discount" bson:"discount"`
	TotalPaid  float64            `json:"total_paid" bson:"total_paid"`
	ReceivedBy primitive.ObjectID `json:"received_by" bson:"received_by"`
}

// Pawn represents a pawn transaction (จำนำทอง)
type Pawn struct {
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	BranchID         primitive.ObjectID `json:"branch_id" bson:"branch_id"`
	PawnNumber       string             `json:"pawn_number" bson:"pawn_number"`
	CustomerID       primitive.ObjectID `json:"customer_id" bson:"customer_id"`
	UserID           primitive.ObjectID `json:"user_id" bson:"user_id"`
	Items            []PawnItem         `json:"items" bson:"items"`
	Principal        float64            `json:"principal" bson:"principal"`         // เงินต้น
	InterestRate     float64            `json:"interest_rate" bson:"interest_rate"` // % per month
	TermMonths       int                `json:"term_months" bson:"term_months"`
	StartDate        time.Time          `json:"start_date" bson:"start_date"`
	DueDate          time.Time          `json:"due_date" bson:"due_date"`
	InterestPayments []InterestPayment  `json:"interest_payments" bson:"interest_payments"`
	Redemption       *Redemption        `json:"redemption,omitempty" bson:"redemption,omitempty"`
	Status           PawnStatus         `json:"status" bson:"status"`
	Notes            string             `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedAt        time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at" bson:"updated_at"`
}

// NewPawn creates a new Pawn entity
func NewPawn(branchID, customerID, userID primitive.ObjectID, pawnNumber string, principal, interestRate float64, termMonths int) *Pawn {
	now := time.Now()
	dueDate := now.AddDate(0, termMonths, 0)

	return &Pawn{
		BranchID:         branchID,
		PawnNumber:       pawnNumber,
		CustomerID:       customerID,
		UserID:           userID,
		Principal:        principal,
		InterestRate:     interestRate,
		TermMonths:       termMonths,
		StartDate:        now,
		DueDate:          dueDate,
		Items:            []PawnItem{},
		InterestPayments: []InterestPayment{},
		Status:           PawnStatusActive,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// AddItem adds an item to the pawn
func (p *Pawn) AddItem(item PawnItem) {
	p.Items = append(p.Items, item)
}

// CalculateInterest calculates interest for a given period
func (p *Pawn) CalculateInterest(months int) float64 {
	return utils.RoundBaht(p.Principal * (p.InterestRate / 100) * float64(months))
}

// CalculateTotalInterestDue calculates total interest due from last payment to now.
// Months are counted in 30-day blocks; partial months accrue zero interest until matured.
func (p *Pawn) CalculateTotalInterestDue() float64 {
	lastPaymentDate := p.StartDate
	if len(p.InterestPayments) > 0 {
		lastPaymentDate = p.InterestPayments[len(p.InterestPayments)-1].PeriodTo
	}

	months := monthsBetween(lastPaymentDate, time.Now())
	return p.CalculateInterest(months)
}

// PayInterest records an interest payment
func (p *Pawn) PayInterest(amount float64, periodFrom, periodTo time.Time, receivedBy primitive.ObjectID) {
	payment := InterestPayment{
		PaymentDate: time.Now(),
		Amount:      amount,
		PeriodFrom:  periodFrom,
		PeriodTo:    periodTo,
		ReceivedBy:  receivedBy,
	}
	p.InterestPayments = append(p.InterestPayments, payment)
	p.UpdatedAt = time.Now()
}

// Redeem redeems the pawn
func (p *Pawn) Redeem(interest, discount float64, receivedBy primitive.ObjectID) {
	totalPaid := utils.RoundBaht(p.Principal + interest - discount)
	if totalPaid < 0 {
		totalPaid = 0
	}
	p.Redemption = &Redemption{
		Date:       time.Now(),
		Principal:  p.Principal,
		Interest:   utils.RoundBaht(interest),
		Discount:   utils.RoundBaht(discount),
		TotalPaid:  totalPaid,
		ReceivedBy: receivedBy,
	}
	p.Status = PawnStatusRedeemed
	p.UpdatedAt = time.Now()
}

// Extend extends the pawn term
func (p *Pawn) Extend(additionalMonths int) {
	p.DueDate = p.DueDate.AddDate(0, additionalMonths, 0)
	p.TermMonths += additionalMonths
	p.Status = PawnStatusActive
	p.UpdatedAt = time.Now()
}

// Forfeit marks the pawn as forfeited
func (p *Pawn) Forfeit() {
	p.Status = PawnStatusForfeited
	p.UpdatedAt = time.Now()
}

// IsOverdue checks if the pawn is overdue
func (p *Pawn) IsOverdue() bool {
	return time.Now().After(p.DueDate) && p.Status == PawnStatusActive
}

// DaysUntilDue returns days until due date (negative if overdue)
func (p *Pawn) DaysUntilDue() int {
	return int(time.Until(p.DueDate).Hours() / 24)
}

// monthsBetween counts elapsed 30-day periods between two timestamps.
// Returns 0 when end is not after start, or when the gap is shorter than one
// matured 30-day period — partial months accrue no interest.
func monthsBetween(start, end time.Time) int {
	if !end.After(start) {
		return 0
	}
	days := int(end.Sub(start).Hours() / 24)
	return days / 30
}
