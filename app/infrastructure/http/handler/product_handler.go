package handler

import (
	"strconv"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/feature/product"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProductHandler handles product (catalog master) and product-item (physical
// stock) HTTP requests.
type ProductHandler struct {
	service *product.Service
}

func NewProductHandler(service *product.Service) *ProductHandler {
	return &ProductHandler{service: service}
}

// ── Master CRUD ───────────────────────────────────────────────────────────────

// CreateProductRequest covers both kinds; only the kind-relevant fields are
// inspected (validate at service).
type CreateProductRequest struct {
	SKU              string   `json:"sku" binding:"required"`
	Code             string   `json:"code"`
	Kind             string   `json:"kind" binding:"required,oneof=ornament bar"`
	GoldType         string   `json:"gold_type" binding:"required"`
	Name             string   `json:"name" binding:"required"`
	Description      string   `json:"description"`
	Design           string   `json:"design"`
	BarSizeBaht      *float64 `json:"bar_size_baht"`
	DefaultLaborCost float64  `json:"default_labor_cost"`
	Images           []string `json:"images"`
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "PRD-400-001", err.Error())
		return
	}
	branchID, _ := primitive.ObjectIDFromHex(c.GetString("BranchId"))
	created, err := h.service.CreateProduct(c.Request.Context(), product.CreateProductInput{
		BranchID:         branchID,
		SKU:              req.SKU,
		Code:             req.Code,
		Kind:             entity.ProductKind(req.Kind),
		GoldType:         req.GoldType,
		Name:             req.Name,
		Description:      req.Description,
		Design:           req.Design,
		BarSizeBaht:      req.BarSizeBaht,
		DefaultLaborCost: req.DefaultLaborCost,
		Images:           req.Images,
	})
	if err != nil {
		utils.BadRequestResponse(c, "PRD-400-002", err.Error())
		return
	}
	utils.CreatedResponse(c, created)
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "PRD-400-003", "invalid product ID")
		return
	}
	p, err := h.service.GetProduct(c.Request.Context(), id)
	if err != nil || p == nil {
		utils.NotFoundResponse(c, "PRD-404-001", "product not found")
		return
	}
	utils.SuccessResponse(c, p)
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	branchID, _ := primitive.ObjectIDFromHex(c.GetString("BranchId"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	products, err := h.service.ListProducts(c.Request.Context(), product.ListProductsInput{
		BranchID: branchID,
		Kind:     entity.ProductKind(c.Query("kind")),
		Search:   c.Query("search"),
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		utils.InternalErrorResponse(c, "PRD-500-001", err.Error())
		return
	}
	utils.SuccessResponse(c, products)
}

type UpdateProductRequest struct {
	Name             *string  `json:"name"`
	Description      *string  `json:"description"`
	Design           *string  `json:"design"`
	DefaultLaborCost *float64 `json:"default_labor_cost"`
	BarSizeBaht      *float64 `json:"bar_size_baht"`
	Images           []string `json:"images"`
	IsActive         *bool    `json:"is_active"`
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "PRD-400-004", "invalid product ID")
		return
	}
	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "PRD-400-005", err.Error())
		return
	}
	p, err := h.service.UpdateProduct(c.Request.Context(), id, product.UpdateProductInput{
		Name:             req.Name,
		Description:      req.Description,
		Design:           req.Design,
		DefaultLaborCost: req.DefaultLaborCost,
		BarSizeBaht:      req.BarSizeBaht,
		Images:           req.Images,
		IsActive:         req.IsActive,
	})
	if err != nil {
		utils.BadRequestResponse(c, "PRD-400-006", err.Error())
		return
	}
	utils.SuccessResponse(c, p)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "PRD-400-007", "invalid product ID")
		return
	}
	if err := h.service.DeleteProduct(c.Request.Context(), id); err != nil {
		utils.BadRequestResponse(c, "PRD-400-008", err.Error())
		return
	}
	utils.MessageResponse(c, "Product deleted")
}

// ── ProductItem CRUD ──────────────────────────────────────────────────────────

type CreateItemRequest struct {
	Barcode      string   `json:"barcode" binding:"required"`
	SerialNumber string   `json:"serial_number"`
	WeightGrams  float64  `json:"weight_grams"`
	LaborCost    *float64 `json:"labor_cost"`
	Cost         float64  `json:"cost"`
	Note         string   `json:"note"`
}

type BulkCreateItemsRequest struct {
	Count       int      `json:"count" binding:"required,min=1"`
	WeightGrams float64  `json:"weight_grams"`
	LaborCost   *float64 `json:"labor_cost"`
	Cost        float64  `json:"cost"`
	BarcodeSeed string   `json:"barcode_seed"`
}

type UpdateItemRequest struct {
	WeightGrams  *float64 `json:"weight_grams"`
	LaborCost    *float64 `json:"labor_cost"`
	Cost         *float64 `json:"cost"`
	SerialNumber *string  `json:"serial_number"`
	Note         *string  `json:"note"`
}

func (h *ProductHandler) ListItems(c *gin.Context) {
	productID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "PRD-400-009", "invalid product ID")
		return
	}
	var statuses []entity.ProductStatus
	if s := c.Query("status"); s != "" {
		statuses = []entity.ProductStatus{entity.ProductStatus(s)}
	}
	items, err := h.service.ListItems(c.Request.Context(), productID, statuses)
	if err != nil {
		utils.InternalErrorResponse(c, "PRD-500-002", err.Error())
		return
	}
	utils.SuccessResponse(c, items)
}

func (h *ProductHandler) CreateItem(c *gin.Context) {
	productID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "PRD-400-010", "invalid product ID")
		return
	}
	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "PRD-400-011", err.Error())
		return
	}
	by, _ := primitive.ObjectIDFromHex(c.GetString("UserId"))
	item, err := h.service.CreateItem(c.Request.Context(), product.CreateItemInput{
		ProductID:    productID,
		Barcode:      req.Barcode,
		SerialNumber: req.SerialNumber,
		WeightGrams:  req.WeightGrams,
		LaborCost:    req.LaborCost,
		Cost:         req.Cost,
		Note:         req.Note,
		By:           by,
	})
	if err != nil {
		utils.BadRequestResponse(c, "PRD-400-012", err.Error())
		return
	}
	utils.CreatedResponse(c, item)
}

func (h *ProductHandler) BulkCreateItems(c *gin.Context) {
	productID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "PRD-400-013", "invalid product ID")
		return
	}
	var req BulkCreateItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "PRD-400-014", err.Error())
		return
	}
	by, _ := primitive.ObjectIDFromHex(c.GetString("UserId"))
	items, err := h.service.BulkCreateItems(c.Request.Context(), product.BulkCreateItemsInput{
		ProductID:   productID,
		Count:       req.Count,
		WeightGrams: req.WeightGrams,
		LaborCost:   req.LaborCost,
		Cost:        req.Cost,
		BarcodeSeed: req.BarcodeSeed,
		By:          by,
	})
	if err != nil {
		utils.BadRequestResponse(c, "PRD-400-015", err.Error())
		return
	}
	utils.CreatedResponse(c, items)
}

func (h *ProductHandler) UpdateItem(c *gin.Context) {
	itemID, err := primitive.ObjectIDFromHex(c.Param("itemId"))
	if err != nil {
		utils.BadRequestResponse(c, "PRD-400-016", "invalid item ID")
		return
	}
	var req UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "PRD-400-017", err.Error())
		return
	}
	item, err := h.service.UpdateItem(c.Request.Context(), itemID, product.UpdateItemInput{
		WeightGrams:  req.WeightGrams,
		LaborCost:    req.LaborCost,
		Cost:         req.Cost,
		SerialNumber: req.SerialNumber,
		Note:         req.Note,
	})
	if err != nil {
		utils.BadRequestResponse(c, "PRD-400-018", err.Error())
		return
	}
	utils.SuccessResponse(c, item)
}

func (h *ProductHandler) DeleteItem(c *gin.Context) {
	itemID, err := primitive.ObjectIDFromHex(c.Param("itemId"))
	if err != nil {
		utils.BadRequestResponse(c, "PRD-400-019", "invalid item ID")
		return
	}
	by, _ := primitive.ObjectIDFromHex(c.GetString("UserId"))
	if err := h.service.DeleteItem(c.Request.Context(), itemID, by); err != nil {
		utils.BadRequestResponse(c, "PRD-400-020", err.Error())
		return
	}
	utils.MessageResponse(c, "Item deleted")
}
