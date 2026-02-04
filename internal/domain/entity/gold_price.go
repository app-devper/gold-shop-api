package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GoldPrice represents gold price at a specific time
type GoldPrice struct {
	ID               primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	Date             time.Time           `json:"date" bson:"date"`
	GoldBarBuy       float64             `json:"gold_bar_buy" bson:"gold_bar_buy"`
	GoldBarSell      float64             `json:"gold_bar_sell" bson:"gold_bar_sell"`
	GoldOrnamentBuy  float64             `json:"gold_ornament_buy" bson:"gold_ornament_buy"`
	GoldOrnamentSell float64             `json:"gold_ornament_sell" bson:"gold_ornament_sell"`
	Source           string              `json:"source" bson:"source"`
	UpdatedBy        *primitive.ObjectID `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	IsActive         bool                `json:"is_active" bson:"is_active"`
	CreatedAt        time.Time           `json:"created_at" bson:"created_at"`
}

// NewGoldPrice creates a new GoldPrice entity
func NewGoldPrice(barBuy, barSell, ornamentBuy, ornamentSell float64, source string) *GoldPrice {
	now := time.Now()
	return &GoldPrice{
		Date:             now,
		GoldBarBuy:       barBuy,
		GoldBarSell:      barSell,
		GoldOrnamentBuy:  ornamentBuy,
		GoldOrnamentSell: ornamentSell,
		Source:           source,
		IsActive:         true,
		CreatedAt:        now,
	}
}

// SetManualUpdate marks the price as manually updated
func (gp *GoldPrice) SetManualUpdate(userID primitive.ObjectID) {
	gp.UpdatedBy = &userID
	gp.Source = "manual"
}
