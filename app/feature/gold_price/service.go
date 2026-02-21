package gold_price

import (
	"context"
	"errors"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service handles gold price logic
type Service struct {
	goldPriceRepo repository.GoldPriceRepository
	apiClient     ExternalGoldPriceAPI
}

// ExternalGoldPriceAPI defines the interface for external gold price API
type ExternalGoldPriceAPI interface {
	GetCurrentPrice(ctx context.Context) (*GoldPriceData, error)
}

// GoldPriceData represents gold price data from external API
type GoldPriceData struct {
	GoldBarBuy       float64
	GoldBarSell      float64
	GoldOrnamentBuy  float64
	GoldOrnamentSell float64
}

// NewService creates a new GoldPrice service
func NewService(goldPriceRepo repository.GoldPriceRepository, apiClient ExternalGoldPriceAPI) *Service {
	return &Service{
		goldPriceRepo: goldPriceRepo,
		apiClient:     apiClient,
	}
}

// GetCurrent retrieves current gold price
func (s *Service) GetCurrent(ctx context.Context) (*entity.GoldPrice, error) {
	return s.goldPriceRepo.GetCurrent(ctx)
}

// GetHistory retrieves gold price history
func (s *Service) GetHistory(ctx context.Context, limit int) ([]*entity.GoldPrice, error) {
	return s.goldPriceRepo.GetHistory(ctx, limit)
}

// SetPrice sets gold price manually
func (s *Service) SetPrice(ctx context.Context, barBuy, barSell, ornamentBuy, ornamentSell float64, userID primitive.ObjectID) (*entity.GoldPrice, error) {
	// Deactivate all current prices
	if err := s.goldPriceRepo.DeactivateAll(ctx); err != nil {
		return nil, errors.New("failed to deactivate old prices")
	}

	// Create new price
	price := entity.NewGoldPrice(barBuy, barSell, ornamentBuy, ornamentSell, "manual")
	price.SetManualUpdate(userID)

	if err := s.goldPriceRepo.Create(ctx, price); err != nil {
		return nil, errors.New("failed to save gold price")
	}

	return price, nil
}

// SyncFromAPI syncs gold price from external API
func (s *Service) SyncFromAPI(ctx context.Context) (*entity.GoldPrice, error) {
	if s.apiClient == nil {
		return nil, errors.New("external API client not configured")
	}

	data, err := s.apiClient.GetCurrentPrice(ctx)
	if err != nil {
		return nil, errors.New("failed to fetch price from API")
	}

	// Deactivate all current prices
	if err := s.goldPriceRepo.DeactivateAll(ctx); err != nil {
		return nil, errors.New("failed to deactivate old prices")
	}

	// Create new price
	price := entity.NewGoldPrice(
		data.GoldBarBuy,
		data.GoldBarSell,
		data.GoldOrnamentBuy,
		data.GoldOrnamentSell,
		"api",
	)

	if err := s.goldPriceRepo.Create(ctx, price); err != nil {
		return nil, errors.New("failed to save gold price")
	}

	return price, nil
}
