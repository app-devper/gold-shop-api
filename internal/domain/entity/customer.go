package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MembershipTier represents customer membership levels
type MembershipTier string

const (
	TierBronze   MembershipTier = "bronze"
	TierSilver   MembershipTier = "silver"
	TierGold     MembershipTier = "gold"
	TierPlatinum MembershipTier = "platinum"
)

// Membership contains membership information
type Membership struct {
	Tier     MembershipTier `json:"tier" bson:"tier"`
	Points   int            `json:"points" bson:"points"`
	JoinDate time.Time      `json:"join_date" bson:"join_date"`
}

// Customer represents a customer in the system
type Customer struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	MemberCode string             `json:"member_code,omitempty" bson:"member_code,omitempty"`
	RFIDCard   string             `json:"rfid_card,omitempty" bson:"rfid_card,omitempty"`
	FullName   string             `json:"full_name" bson:"full_name"`
	IDCard     string             `json:"id_card,omitempty" bson:"id_card,omitempty"`
	Phone      string             `json:"phone" bson:"phone"`
	Email      string             `json:"email,omitempty" bson:"email,omitempty"`
	Address    string             `json:"address,omitempty" bson:"address,omitempty"`
	IsMember   bool               `json:"is_member" bson:"is_member"`
	Membership *Membership        `json:"membership,omitempty" bson:"membership,omitempty"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}

// NewCustomer creates a new Customer entity
func NewCustomer(fullName, phone string, isMember bool) *Customer {
	now := time.Now()
	customer := &Customer{
		FullName:  fullName,
		Phone:     phone,
		IsMember:  isMember,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if isMember {
		customer.Membership = &Membership{
			Tier:     TierBronze,
			Points:   0,
			JoinDate: now,
		}
	}

	return customer
}

// AddPoints adds points to customer membership
func (c *Customer) AddPoints(points int) {
	if c.Membership != nil {
		c.Membership.Points += points
		c.UpdateMembershipTier()
	}
}

// RedeemPoints redeems points from customer membership
func (c *Customer) RedeemPoints(points int) bool {
	if c.Membership != nil && c.Membership.Points >= points {
		c.Membership.Points -= points
		return true
	}
	return false
}

// UpdateMembershipTier updates tier based on points
func (c *Customer) UpdateMembershipTier() {
	if c.Membership == nil {
		return
	}

	switch {
	case c.Membership.Points >= 100000:
		c.Membership.Tier = TierPlatinum
	case c.Membership.Points >= 50000:
		c.Membership.Tier = TierGold
	case c.Membership.Points >= 20000:
		c.Membership.Tier = TierSilver
	default:
		c.Membership.Tier = TierBronze
	}
}
