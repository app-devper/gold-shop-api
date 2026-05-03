package entity

import (
	"errors"
	"time"

	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GoldSavingStatus represents the status of a gold saving account
type GoldSavingStatus string

const (
	GoldSavingStatusActive GoldSavingStatus = "active"
	GoldSavingStatusClosed GoldSavingStatus = "closed"
)

// TxType — direction of a transaction.
type TxType string

const (
	TxDeposit  TxType = "deposit"
	TxWithdraw TxType = "withdraw"
	TxAdjust   TxType = "adjust" // admin manual correction
)

// TxMode — how the customer transacts.
//
//	cash     — operator enters baht; system derives gold weight using current price.
//	physical — operator enters grams directly; cash equivalent is computed for the audit log.
type TxMode string

const (
	TxModeCash     TxMode = "cash"
	TxModePhysical TxMode = "physical"
)

// GoldSavingTransaction is an append-only ledger entry on the account.
//
// Snapshot semantics:
//   - GoldPricePerGram captures the price actually used for this tx
//     (sell-side for deposits, buy-side for withdraws). Future market moves
//     do not retro-affect history.
//   - GoldWeightDelta is signed: positive on deposit/credit-adjust, negative
//     on withdraw/debit-adjust. Sums of the deltas equal the running balance.
type GoldSavingTransaction struct {
	Date             time.Time          `json:"date" bson:"date"`
	Type             TxType             `json:"type" bson:"type"`
	Mode             TxMode             `json:"mode" bson:"mode"`
	InputAmount      float64            `json:"input_amount" bson:"input_amount"`               // ฿ when Mode=cash, grams when Mode=physical
	GoldPricePerGram float64            `json:"gold_price_per_gram" bson:"gold_price_per_gram"` // snapshot
	GoldWeightDelta  float64            `json:"gold_weight_delta" bson:"gold_weight_delta"`     // signed
	CashEquivalent   float64            `json:"cash_equivalent" bson:"cash_equivalent"`         // ฿ value of the tx (always positive)
	BalanceAfter     float64            `json:"balance_after" bson:"balance_after"`             // grams
	ProcessedBy      primitive.ObjectID `json:"processed_by" bson:"processed_by"`
	Note             string             `json:"note,omitempty" bson:"note,omitempty"`
}

// GoldSaving is a unified single-balance saving account.
// Balance is gold weight (grams). Cash side of any transaction is computed
// against gold price snapshot at tx time and never persisted as standalone state.
type GoldSaving struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AccountNumber string             `json:"account_number" bson:"account_number"`
	BranchID      primitive.ObjectID `json:"branch_id" bson:"branch_id"`
	CustomerID    primitive.ObjectID `json:"customer_id" bson:"customer_id"`

	// Primary balance — grams of gold held for the customer.
	GoldWeight float64 `json:"gold_weight" bson:"gold_weight"`

	// Lifetime aggregates for cost-basis / statement.
	TotalDepositValue   float64 `json:"total_deposit_value" bson:"total_deposit_value"`     // ฿
	TotalDepositWeight  float64 `json:"total_deposit_weight" bson:"total_deposit_weight"`   // g
	TotalWithdrawValue  float64 `json:"total_withdraw_value" bson:"total_withdraw_value"`   // ฿
	TotalWithdrawWeight float64 `json:"total_withdraw_weight" bson:"total_withdraw_weight"` // g

	// Per-mode minimums (admin can leave any as 0 to disable).
	MinDepositCash      float64 `json:"min_deposit_cash" bson:"min_deposit_cash"`           // ฿
	MinDepositPhysical  float64 `json:"min_deposit_physical" bson:"min_deposit_physical"`   // g
	MinWithdrawCash     float64 `json:"min_withdraw_cash" bson:"min_withdraw_cash"`         // ฿
	MinWithdrawPhysical float64 `json:"min_withdraw_physical" bson:"min_withdraw_physical"` // g

	Status     GoldSavingStatus `json:"status" bson:"status"`
	OpenedDate time.Time        `json:"opened_date" bson:"opened_date"`
	ClosedDate *time.Time       `json:"closed_date,omitempty" bson:"closed_date,omitempty"`

	Transactions []GoldSavingTransaction `json:"transactions" bson:"transactions"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// NewGoldSaving creates a fresh active account with zero balances.
func NewGoldSaving(branchID, customerID primitive.ObjectID, accountNumber string) *GoldSaving {
	now := time.Now()
	return &GoldSaving{
		BranchID:      branchID,
		CustomerID:    customerID,
		AccountNumber: accountNumber,
		Status:        GoldSavingStatusActive,
		OpenedDate:    now,
		Transactions:  []GoldSavingTransaction{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// SetMinimums applies optional per-mode minimum thresholds. Negatives are coerced to 0.
func (gs *GoldSaving) SetMinimums(depCash, depGold, wdCash, wdGold float64) {
	gs.MinDepositCash = nonNeg(depCash)
	gs.MinDepositPhysical = nonNeg(depGold)
	gs.MinWithdrawCash = nonNeg(wdCash)
	gs.MinWithdrawPhysical = nonNeg(wdGold)
}

// DepositCash credits the account using a baht amount. Weight is derived from
// the sell-side per-gram price and stored along with the snapshot.
func (gs *GoldSaving) DepositCash(amountBaht, sellPricePerGram float64, by primitive.ObjectID) error {
	if amountBaht <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if sellPricePerGram <= 0 {
		return errors.New("invalid gold price")
	}
	if gs.MinDepositCash > 0 && amountBaht < gs.MinDepositCash {
		return errors.New("amount is below minimum cash deposit")
	}

	weight := utils.RoundGram(amountBaht / sellPricePerGram)
	gs.applyDelta(weight, amountBaht, GoldSavingTransaction{
		Type:             TxDeposit,
		Mode:             TxModeCash,
		InputAmount:      amountBaht,
		GoldPricePerGram: sellPricePerGram,
		GoldWeightDelta:  weight,
		CashEquivalent:   utils.RoundBaht(amountBaht),
		ProcessedBy:      by,
	})
	gs.TotalDepositValue = utils.RoundBaht(gs.TotalDepositValue + amountBaht)
	gs.TotalDepositWeight = utils.RoundGram(gs.TotalDepositWeight + weight)
	return nil
}

// DepositGold credits the account by physical weight. Cash equivalent is
// computed at sell-side price for cost-basis tracking only.
func (gs *GoldSaving) DepositGold(weight, sellPricePerGram float64, by primitive.ObjectID) error {
	if weight <= 0 {
		return errors.New("weight must be greater than zero")
	}
	if sellPricePerGram <= 0 {
		return errors.New("invalid gold price")
	}
	if gs.MinDepositPhysical > 0 && weight < gs.MinDepositPhysical {
		return errors.New("weight is below minimum physical deposit")
	}

	weight = utils.RoundGram(weight)
	cash := utils.RoundBaht(weight * sellPricePerGram)
	gs.applyDelta(weight, cash, GoldSavingTransaction{
		Type:             TxDeposit,
		Mode:             TxModePhysical,
		InputAmount:      weight,
		GoldPricePerGram: sellPricePerGram,
		GoldWeightDelta:  weight,
		CashEquivalent:   cash,
		ProcessedBy:      by,
	})
	gs.TotalDepositValue = utils.RoundBaht(gs.TotalDepositValue + cash)
	gs.TotalDepositWeight = utils.RoundGram(gs.TotalDepositWeight + weight)
	return nil
}

// WithdrawCash debits the account by a baht cashout. Weight is derived from
// the buy-side per-gram price (mark-to-market). Returns ErrInsufficientBalance
// if the resulting weight would push the balance negative.
func (gs *GoldSaving) WithdrawCash(amountBaht, buyPricePerGram float64, by primitive.ObjectID) error {
	if amountBaht <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if buyPricePerGram <= 0 {
		return errors.New("invalid gold price")
	}
	if gs.MinWithdrawCash > 0 && amountBaht < gs.MinWithdrawCash {
		return errors.New("amount is below minimum cash withdrawal")
	}

	weight := utils.RoundGram(amountBaht / buyPricePerGram)
	if weight > gs.GoldWeight {
		return ErrInsufficientBalance
	}

	gs.applyDelta(-weight, amountBaht, GoldSavingTransaction{
		Type:             TxWithdraw,
		Mode:             TxModeCash,
		InputAmount:      amountBaht,
		GoldPricePerGram: buyPricePerGram,
		GoldWeightDelta:  -weight,
		CashEquivalent:   utils.RoundBaht(amountBaht),
		ProcessedBy:      by,
	})
	gs.TotalWithdrawValue = utils.RoundBaht(gs.TotalWithdrawValue + amountBaht)
	gs.TotalWithdrawWeight = utils.RoundGram(gs.TotalWithdrawWeight + weight)
	return nil
}

// WithdrawGold debits the account by physical weight (customer collects gold).
func (gs *GoldSaving) WithdrawGold(weight, buyPricePerGram float64, by primitive.ObjectID) error {
	if weight <= 0 {
		return errors.New("weight must be greater than zero")
	}
	if buyPricePerGram <= 0 {
		return errors.New("invalid gold price")
	}
	if gs.MinWithdrawPhysical > 0 && weight < gs.MinWithdrawPhysical {
		return errors.New("weight is below minimum physical withdrawal")
	}

	weight = utils.RoundGram(weight)
	if weight > gs.GoldWeight {
		return ErrInsufficientBalance
	}
	cash := utils.RoundBaht(weight * buyPricePerGram)
	gs.applyDelta(-weight, cash, GoldSavingTransaction{
		Type:             TxWithdraw,
		Mode:             TxModePhysical,
		InputAmount:      weight,
		GoldPricePerGram: buyPricePerGram,
		GoldWeightDelta:  -weight,
		CashEquivalent:   cash,
		ProcessedBy:      by,
	})
	gs.TotalWithdrawValue = utils.RoundBaht(gs.TotalWithdrawValue + cash)
	gs.TotalWithdrawWeight = utils.RoundGram(gs.TotalWithdrawWeight + weight)
	return nil
}

// Adjust applies an admin-driven correction (signed weight delta in grams).
// The optional currentPrice is used to compute a cash equivalent for the
// audit entry; pass 0 if no price is available. Note is required by the
// service layer (entity layer just records it).
func (gs *GoldSaving) Adjust(weightDelta, currentPrice float64, note string, by primitive.ObjectID) error {
	if weightDelta == 0 {
		return errors.New("adjustment delta must be non-zero")
	}
	weightDelta = utils.RoundGram(weightDelta)
	if weightDelta < 0 && -weightDelta > gs.GoldWeight {
		return ErrInsufficientBalance
	}
	cash := 0.0
	if currentPrice > 0 {
		cash = utils.RoundBaht(weightDelta * currentPrice)
		if cash < 0 {
			cash = -cash
		}
	}
	gs.applyDelta(weightDelta, cash, GoldSavingTransaction{
		Type:             TxAdjust,
		Mode:             TxModePhysical,
		InputAmount:      weightDelta,
		GoldPricePerGram: currentPrice,
		GoldWeightDelta:  weightDelta,
		CashEquivalent:   cash,
		ProcessedBy:      by,
		Note:             note,
	})
	return nil
}

// CurrentValue returns gs.GoldWeight × currentBuyPricePerGram (mark-to-market).
func (gs *GoldSaving) CurrentValue(currentBuyPricePerGram float64) float64 {
	return utils.RoundBaht(gs.GoldWeight * currentBuyPricePerGram)
}

// CostBasis is total deposit value − total withdraw value (running net cost).
// Negative when the customer has withdrawn more cash than they have deposited.
func (gs *GoldSaving) CostBasis() float64 {
	return utils.RoundBaht(gs.TotalDepositValue - gs.TotalWithdrawValue)
}

// Close closes the account. Service layer is responsible for ensuring the
// balance is at/near zero before calling.
func (gs *GoldSaving) Close() {
	now := time.Now()
	gs.Status = GoldSavingStatusClosed
	gs.ClosedDate = &now
	gs.UpdatedAt = now
}

// applyDelta is the only mutator of GoldWeight + Transactions.
// `cash` is the |baht| value associated with the entry (for the audit row).
func (gs *GoldSaving) applyDelta(weightDelta, _cash float64, tx GoldSavingTransaction) {
	gs.GoldWeight = utils.RoundGram(gs.GoldWeight + weightDelta)
	tx.Date = time.Now()
	tx.BalanceAfter = gs.GoldWeight
	gs.Transactions = append(gs.Transactions, tx)
	gs.UpdatedAt = tx.Date
	_ = _cash // value is already on the tx itself; kept for symmetry / future hooks
}

func nonNeg(v float64) float64 {
	if v < 0 {
		return 0
	}
	return v
}
