package handler

import (
	"strconv"

	"github.com/devper-gold/gold-shop-api/internal/application/sale"
	"github.com/devper-gold/gold-shop-api/internal/domain/entity"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SaleHandler handles sale endpoints
type SaleHandler struct {
	saleService *sale.Service
}

// NewSaleHandler creates a new SaleHandler
func NewSaleHandler(saleService *sale.Service) *SaleHandler {
	return &SaleHandler{
		saleService: saleService,
	}
}

// CreateSaleRequest represents create sale request
type CreateSaleRequest struct {
	CustomerID   string            `json:"customer_id"`
	SaleType     string            `json:"sale_type" binding:"required"`
	Items        []SaleItemRequest `json:"items" binding:"required"`
	OldGoldItems []OldGoldRequest  `json:"old_gold_items"`
	Discount     float64           `json:"discount"`
	DiscountType string            `json:"discount_type"`
	Payments     []PaymentRequest  `json:"payments" binding:"required"`
	PointsUsed   int               `json:"points_used"`
	Notes        string            `json:"notes"`
}

// SaleItemRequest represents a sale item in request
type SaleItemRequest struct {
	ProductID     string   `json:"product_id" binding:"required"`
	ProductItemID string   `json:"product_item_id"`
	PriceLevel    string   `json:"price_level"`
	Price         *float64 `json:"price,omitempty"`
	Weight        float64  `json:"weight"`
	Discount      float64  `json:"discount"`
	DiscountType  string   `json:"discount_type"`
}

// OldGoldRequest represents old gold in request
type OldGoldRequest struct {
	Description  string  `json:"description"`
	GoldType     string  `json:"gold_type"`
	Weight       float64 `json:"weight"`
	PricePerUnit float64 `json:"price_per_unit"`
}

// PaymentRequest represents payment in request
type PaymentRequest struct {
	Method    string  `json:"method" binding:"required"`
	Amount    float64 `json:"amount" binding:"required"`
	Reference string  `json:"reference"`
}

// List returns sales with filters
func (h *SaleHandler) List(c *gin.Context) {
	branchID, _ := primitive.ObjectIDFromHex(c.Query("branch_id"))
	if branchID.IsZero() {
		branchID, _ = primitive.ObjectIDFromHex(c.GetString("branch_id"))
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var status []entity.SaleStatus
	if s := c.Query("status"); s != "" {
		status = []entity.SaleStatus{entity.SaleStatus(s)}
	}

	sales, err := h.saleService.GetByBranchID(c.Request.Context(), branchID, status, limit, offset)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch sales")
		return
	}

	utils.SuccessResponse(c, sales)
}

// Get returns a sale by ID
func (h *SaleHandler) Get(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid sale ID")
		return
	}

	saleEntity, err := h.saleService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == entity.ErrNotFound {
			utils.NotFoundResponse(c, "Sale not found")
			return
		}
		utils.InternalErrorResponse(c, "Failed to fetch sale")
		return
	}

	utils.SuccessResponse(c, saleEntity)
}

// Create creates a new sale
func (h *SaleHandler) Create(c *gin.Context) {
	var req CreateSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	branchID, _ := primitive.ObjectIDFromHex(c.GetString("branch_id"))
	userID, _ := primitive.ObjectIDFromHex(c.GetString("user_id"))

	input := sale.CreateSaleInput{
		BranchID:     branchID,
		UserID:       userID,
		CustomerID:   req.CustomerID,
		SaleType:     entity.SaleType(req.SaleType),
		Discount:     req.Discount,
		DiscountType: entity.DiscountType(req.DiscountType),
		PointsUsed:   req.PointsUsed,
		Notes:        req.Notes,
	}

	// Convert items
	for _, item := range req.Items {
		input.Items = append(input.Items, sale.SaleItemInput{
			ProductID:     item.ProductID,
			ProductItemID: item.ProductItemID,
			PriceLevel:    item.PriceLevel,
			Price:         item.Price,
			Weight:        item.Weight,
			Discount:      item.Discount,
			DiscountType:  entity.DiscountType(item.DiscountType),
		})
	}

	// Convert old gold items
	for _, item := range req.OldGoldItems {
		input.OldGoldItems = append(input.OldGoldItems, sale.OldGoldInput{
			Description:  item.Description,
			GoldType:     item.GoldType,
			Weight:       item.Weight,
			PricePerUnit: item.PricePerUnit,
		})
	}

	// Convert payments
	for _, p := range req.Payments {
		input.Payments = append(input.Payments, sale.PaymentInput{
			Method:    entity.PaymentMethod(p.Method),
			Amount:    p.Amount,
			Reference: p.Reference,
		})
	}

	saleEntity, err := h.saleService.Create(c.Request.Context(), input)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.CreatedResponse(c, saleEntity)
}

// Cancel cancels a sale
func (h *SaleHandler) Cancel(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid sale ID")
		return
	}

	if err := h.saleService.Cancel(c.Request.Context(), id); err != nil {
		if err == entity.ErrNotFound {
			utils.NotFoundResponse(c, "Sale not found")
			return
		}
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.MessageResponse(c, "Sale cancelled successfully")
}

// GetReceipt returns sale receipt
func (h *SaleHandler) GetReceipt(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid sale ID")
		return
	}

	receipt, err := h.saleService.GenerateReceipt(c.Request.Context(), id)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to generate receipt")
		return
	}

	utils.SuccessResponse(c, receipt)
}

// GetUnpaid returns unpaid sales for the branch
func (h *SaleHandler) GetUnpaid(c *gin.Context) {
	branchID, _ := primitive.ObjectIDFromHex(c.GetString("branch_id"))
	if branchID.IsZero() {
		branchID, _ = primitive.ObjectIDFromHex(c.Query("branch_id"))
	}

	sales, err := h.saleService.GetUnpaidSales(c.Request.Context(), branchID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch unpaid sales")
		return
	}

	utils.SuccessResponse(c, sales)
}
