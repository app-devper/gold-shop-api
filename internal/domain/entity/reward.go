package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Reward represents a redeemable reward
type Reward struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Code           string             `json:"code" bson:"code"`
	Name           string             `json:"name" bson:"name"`
	Description    string             `json:"description" bson:"description"`
	PointsRequired int                `json:"points_required" bson:"points_required"`
	Quantity       int                `json:"quantity" bson:"quantity"`
	Images         []string           `json:"images" bson:"images"`
	IsActive       bool               `json:"is_active" bson:"is_active"`
	ValidFrom      time.Time          `json:"valid_from" bson:"valid_from"`
	ValidUntil     time.Time          `json:"valid_until" bson:"valid_until"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
}

// RewardRedemption represents a reward redemption record
type RewardRedemption struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CustomerID  primitive.ObjectID `json:"customer_id" bson:"customer_id"`
	RewardID    primitive.ObjectID `json:"reward_id" bson:"reward_id"`
	BranchID    primitive.ObjectID `json:"branch_id" bson:"branch_id"`
	PointsUsed  int                `json:"points_used" bson:"points_used"`
	RedeemedAt  time.Time          `json:"redeemed_at" bson:"redeemed_at"`
	ProcessedBy primitive.ObjectID `json:"processed_by" bson:"processed_by"`
}

// NewReward creates a new Reward entity
func NewReward(code, name, description string, pointsRequired, quantity int, validFrom, validUntil time.Time) *Reward {
	return &Reward{
		Code:           code,
		Name:           name,
		Description:    description,
		PointsRequired: pointsRequired,
		Quantity:       quantity,
		Images:         []string{},
		IsActive:       true,
		ValidFrom:      validFrom,
		ValidUntil:     validUntil,
		CreatedAt:      time.Now(),
	}
}

// IsValid checks if reward is currently valid
func (r *Reward) IsValid() bool {
	now := time.Now()
	return r.IsActive && r.Quantity > 0 && now.After(r.ValidFrom) && now.Before(r.ValidUntil)
}

// DeductQuantity deducts quantity after redemption
func (r *Reward) DeductQuantity() {
	if r.Quantity > 0 {
		r.Quantity--
	}
}
