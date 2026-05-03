package entity

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCalculateTotals_NetTotalClampedToZero(t *testing.T) {
	s := NewSale(primitive.NewObjectID(), primitive.NewObjectID(), "S-1", SaleTypeSell)
	s.AddItem(SaleItem{Total: 1000})
	s.Discount = 500
	s.OldGoldValue = 0
	s.AddOldGoldItem(OldGoldItem{Total: 2000}) // forces NetTotal negative pre-clamp
	s.DiscountType = DiscountTypeAmount
	s.CalculateTotals()

	if s.NetTotal < 0 {
		t.Fatalf("NetTotal must not be negative, got %v", s.NetTotal)
	}
	if s.NetTotal != 0 {
		t.Fatalf("NetTotal expected 0 when discount+oldGold exceeds subtotal, got %v", s.NetTotal)
	}
}

func TestCalculateTotals_AppliesPercentDiscountOnSubtotal(t *testing.T) {
	s := NewSale(primitive.NewObjectID(), primitive.NewObjectID(), "S-1", SaleTypeSell)
	s.AddItem(SaleItem{Total: 2000})
	s.Discount = 10
	s.DiscountType = DiscountTypePercent
	s.CalculateTotals()

	if s.NetTotal != 1800 {
		t.Fatalf("expected 1800 after 10%% discount, got %v", s.NetTotal)
	}
}
