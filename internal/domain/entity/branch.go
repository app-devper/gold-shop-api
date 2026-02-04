package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Branch represents a gold shop branch
type Branch struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Code      string             `json:"code" bson:"code"`
	Name      string             `json:"name" bson:"name"`
	Address   string             `json:"address" bson:"address"`
	Phone     string             `json:"phone" bson:"phone"`
	IsActive  bool               `json:"is_active" bson:"is_active"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// NewBranch creates a new Branch entity
func NewBranch(code, name, address, phone string) *Branch {
	now := time.Now()
	return &Branch{
		Code:      code,
		Name:      name,
		Address:   address,
		Phone:     phone,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
