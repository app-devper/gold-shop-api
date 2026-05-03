package pricing

import (
	"testing"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestCalculateSellLine_OrnamentWithLaborAndDiscount(t *testing.T) {
	price := &entity.GoldPrice{GoldOrnamentSell: 30500}
	product := &entity.Product{Kind: entity.KindOrnament}

	quote, err := CalculateSellLine(price, product, 7.6, 500, 100, entity.DiscountTypeAmount, nil)

	assert.NoError(t, err)
	assert.InDelta(t, 2011.87, quote.PricePerGram, 0.01)
	assert.InDelta(t, 15290.24, quote.GoldValue, 0.01)
	assert.Equal(t, 500.0, quote.LaborCost)
	assert.Equal(t, 100.0, quote.Discount)
	assert.InDelta(t, 15690.24, quote.Total, 0.01)
}

func TestCalculateBuyback_AppliesDeductionPercent(t *testing.T) {
	price := &entity.GoldPrice{GoldOrnamentBuy: 29000}

	quote, err := CalculateBuyback(price, entity.KindOrnament, 7.6, 3, 0)

	assert.NoError(t, err)
	assert.InDelta(t, 1912.93, quote.PricePerGram, 0.01)
	assert.InDelta(t, 14538.26, quote.GrossTotal, 0.01)
	assert.InDelta(t, 436.15, quote.DeductionAmount, 0.01)
	assert.InDelta(t, 14102.11, quote.NetTotal, 0.01)
}
