package handler

import (
	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/feature/gold_price"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GoldPriceHandler handles gold price endpoints
type GoldPriceHandler struct {
	goldPriceService *gold_price.Service
}

// NewGoldPriceHandler creates a new GoldPriceHandler
func NewGoldPriceHandler(goldPriceService *gold_price.Service) *GoldPriceHandler {
	return &GoldPriceHandler{
		goldPriceService: goldPriceService,
	}
}

// SetPriceRequest represents set price request
type SetPriceRequest struct {
	GoldBarBuy       float64 `json:"gold_bar_buy" binding:"required"`
	GoldBarSell      float64 `json:"gold_bar_sell" binding:"required"`
	GoldOrnamentBuy  float64 `json:"gold_ornament_buy" binding:"required"`
	GoldOrnamentSell float64 `json:"gold_ornament_sell" binding:"required"`
}

// GetCurrent returns current gold price
func (h *GoldPriceHandler) GetCurrent(c *gin.Context) {
	price, err := h.goldPriceService.GetCurrent(c.Request.Context())
	if err != nil {
		if err == entity.ErrNotFound {
			utils.NotFoundResponse(c, "GLP-404-001", "No gold price set")
			return
		}
		utils.InternalErrorResponse(c, "GLP-500-001", "Failed to fetch gold price")
		return
	}

	utils.SuccessResponse(c, price)
}

// GetHistory returns gold price history
func (h *GoldPriceHandler) GetHistory(c *gin.Context) {
	prices, err := h.goldPriceService.GetHistory(c.Request.Context(), 30)
	if err != nil {
		utils.InternalErrorResponse(c, "GLP-500-002", "Failed to fetch price history")
		return
	}

	utils.SuccessResponse(c, prices)
}

// SetPrice sets gold price manually
func (h *GoldPriceHandler) SetPrice(c *gin.Context) {
	var req SetPriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "GLP-400-001", "Invalid request body")
		return
	}

	userID, _ := primitive.ObjectIDFromHex(c.GetString("UserId"))

	price, err := h.goldPriceService.SetPrice(
		c.Request.Context(),
		req.GoldBarBuy,
		req.GoldBarSell,
		req.GoldOrnamentBuy,
		req.GoldOrnamentSell,
		userID,
	)
	if err != nil {
		utils.InternalErrorResponse(c, "GLP-500-003", "Failed to set gold price")
		return
	}

	utils.CreatedResponse(c, price)
}

// SyncFromAPI syncs gold price from external API
func (h *GoldPriceHandler) SyncFromAPI(c *gin.Context) {
	price, err := h.goldPriceService.SyncFromAPI(c.Request.Context())
	if err != nil {
		utils.InternalErrorResponse(c, "GLP-500-004", "Failed to sync gold price from API")
		return
	}

	utils.SuccessResponse(c, price)
}
