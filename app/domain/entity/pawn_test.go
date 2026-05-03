package entity

import (
	"testing"
	"time"
)

func TestMonthsBetween(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		end  time.Time
		want int
	}{
		{"same instant", base, 0},
		{"end before start", base.AddDate(0, 0, -5), 0},
		{"15 days", base.AddDate(0, 0, 15), 0},
		{"29 days", base.AddDate(0, 0, 29), 0},
		{"30 days", base.AddDate(0, 0, 30), 1},
		{"31 days", base.AddDate(0, 0, 31), 1},
		{"45 days", base.AddDate(0, 0, 45), 1},
		{"60 days", base.AddDate(0, 0, 60), 2},
		{"90 days", base.AddDate(0, 0, 90), 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := monthsBetween(base, tc.end)
			if got != tc.want {
				t.Fatalf("monthsBetween(%v) = %d, want %d", tc.end.Sub(base), got, tc.want)
			}
		})
	}
}

func TestCalculateTotalInterestDue_PartialMonthAccruesZero(t *testing.T) {
	p := &Pawn{
		Principal:    10000,
		InterestRate: 2,
		StartDate:    time.Now().AddDate(0, 0, -5),
	}

	if got := p.CalculateTotalInterestDue(); got != 0 {
		t.Fatalf("expected 0 interest in first 5 days, got %v", got)
	}
}
