package pricing

import (
	"errors"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
)

// SellLineQuote is the canonical price breakdown for one sellable gold item.
type SellLineQuote struct {
	PricePerGram float64
	GoldValue    float64
	LaborCost    float64
	Discount     float64
	Total        float64
}

// BuybackQuote is the canonical price breakdown for old-gold buyback.
type BuybackQuote struct {
	PricePerGram     float64
	GrossTotal       float64
	DeductionPercent float64
	DeductionAmount  float64
	NetTotal         float64
}

// CalculateSellLine calculates a sale line from the current gold price,
// physical item weight, labor fee, and discount policy.
func CalculateSellLine(
	goldPrice *entity.GoldPrice,
	product *entity.Product,
	weightGrams float64,
	laborCost float64,
	discount float64,
	discountType entity.DiscountType,
	manualGoldValue *float64,
) (SellLineQuote, error) {
	if goldPrice == nil {
		return SellLineQuote{}, errors.New("gold price is required")
	}
	if product == nil {
		return SellLineQuote{}, errors.New("product is required")
	}
	if weightGrams <= 0 {
		return SellLineQuote{}, errors.New("weight must be greater than zero")
	}
	if laborCost < 0 {
		return SellLineQuote{}, errors.New("labor cost cannot be negative")
	}
	if discount < 0 {
		return SellLineQuote{}, errors.New("discount cannot be negative")
	}

	pricePerBaht := goldPrice.GoldOrnamentSell
	if product.IsBarGold() {
		pricePerBaht = goldPrice.GoldBarSell
	}
	pricePerGram := pricePerBaht / entity.BahtPerGramFor(product.Kind)
	goldValue := pricePerGram * weightGrams
	if manualGoldValue != nil {
		if *manualGoldValue < 0 {
			return SellLineQuote{}, errors.New("manual price cannot be negative")
		}
		goldValue = *manualGoldValue
		pricePerGram = goldValue / weightGrams
	}

	totalBeforeDiscount := goldValue + laborCost
	discountAmount := discount
	if discountType == entity.DiscountTypePercent {
		discountAmount = totalBeforeDiscount * (discount / 100)
	}
	if discountAmount > totalBeforeDiscount {
		discountAmount = totalBeforeDiscount
	}

	return SellLineQuote{
		PricePerGram: utils.RoundBaht(pricePerGram),
		GoldValue:    utils.RoundBaht(goldValue),
		LaborCost:    utils.RoundBaht(laborCost),
		Discount:     utils.RoundBaht(discountAmount),
		Total:        utils.RoundBaht(totalBeforeDiscount - discountAmount),
	}, nil
}

// CalculateBuyback calculates old-gold buyback from the active gold buy price.
// If manualPricePerGram is positive, it is used as the base buy price per gram.
func CalculateBuyback(
	goldPrice *entity.GoldPrice,
	kind entity.ProductKind,
	weightGrams float64,
	deductionPercent float64,
	manualPricePerGram float64,
) (BuybackQuote, error) {
	if goldPrice == nil {
		return BuybackQuote{}, errors.New("gold price is required")
	}
	if weightGrams <= 0 {
		return BuybackQuote{}, errors.New("weight must be greater than zero")
	}
	if deductionPercent < 0 || deductionPercent > 100 {
		return BuybackQuote{}, errors.New("deduction percent must be between 0 and 100")
	}

	pricePerBaht := goldPrice.GoldOrnamentBuy
	if kind == entity.KindBar {
		pricePerBaht = goldPrice.GoldBarBuy
	}
	pricePerGram := pricePerBaht / entity.BahtPerGramFor(kind)
	if manualPricePerGram > 0 {
		pricePerGram = manualPricePerGram
	}

	grossTotal := pricePerGram * weightGrams
	deductionAmount := grossTotal * (deductionPercent / 100)
	netTotal := grossTotal - deductionAmount
	if netTotal < 0 {
		netTotal = 0
	}

	return BuybackQuote{
		PricePerGram:     utils.RoundBaht(pricePerGram),
		GrossTotal:       utils.RoundBaht(grossTotal),
		DeductionPercent: deductionPercent,
		DeductionAmount:  utils.RoundBaht(deductionAmount),
		NetTotal:         utils.RoundBaht(netTotal),
	}, nil
}
