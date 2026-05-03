package entity

import (
	"math"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestWithdraw_AsCash_ReducesCashBalance(t *testing.T) {
	gs := NewGoldSaving(primitive.NewObjectID(), primitive.NewObjectID(), "GS-1", GoldSavingByMoney, 0, 0)
	user := primitive.NewObjectID()

	if err := gs.Deposit(3000, 30000, user); err != nil {
		t.Fatalf("deposit: %v", err)
	}
	// 3000 / 30000 = 0.1g
	if math.Abs(gs.GoldBalance-0.1) > 1e-6 {
		t.Fatalf("after deposit GoldBalance expected 0.1, got %v", gs.GoldBalance)
	}
	if gs.CashBalance != 3000 {
		t.Fatalf("after deposit CashBalance expected 3000, got %v", gs.CashBalance)
	}

	if err := gs.Withdraw(1500, true, 30000, user); err != nil {
		t.Fatalf("withdraw: %v", err)
	}
	if gs.CashBalance != 1500 {
		t.Fatalf("CashBalance after ฿1500 cash withdraw expected 1500, got %v", gs.CashBalance)
	}
	// 1500 / 30000 = 0.05g remaining = 0.05g
	if math.Abs(gs.GoldBalance-0.05) > 1e-6 {
		t.Fatalf("GoldBalance after withdraw expected 0.05, got %v", gs.GoldBalance)
	}
}

func TestWithdraw_AsGold_ReducesCashBalanceByEquivalent(t *testing.T) {
	gs := NewGoldSaving(primitive.NewObjectID(), primitive.NewObjectID(), "GS-1", GoldSavingByMoney, 0, 0)
	user := primitive.NewObjectID()

	_ = gs.Deposit(3000, 30000, user)
	// withdraw 0.05g physical at price 30000 → equivalent ฿1500
	if err := gs.Withdraw(0.05, false, 30000, user); err != nil {
		t.Fatalf("withdraw gold: %v", err)
	}
	if gs.CashBalance != 1500 {
		t.Fatalf("CashBalance after physical gold withdraw expected 1500, got %v", gs.CashBalance)
	}
}

func TestWithdraw_InsufficientBalance(t *testing.T) {
	gs := NewGoldSaving(primitive.NewObjectID(), primitive.NewObjectID(), "GS-1", GoldSavingByWeight, 0, 0)
	user := primitive.NewObjectID()

	_ = gs.Deposit(0.05, 0, user) // direct gram deposit
	if err := gs.Withdraw(1, false, 30000, user); err == nil {
		t.Fatal("expected ErrInsufficientBalance, got nil")
	}
}
