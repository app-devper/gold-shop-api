package handler

import (
	"strconv"

	"github.com/devper-gold/gold-shop-api/internal/application/product"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
)

// ProductHandler handles product HTTP requests
type ProductHandler struct {
	service *product.Service
}

// NewProductHandler creates a new Product handler
func NewProductHandler(service *product.Service) *ProductHandler {
	return &ProductHandler{service: service}
}

// CreateProduct handles product creation
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req product.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	product, err := h.service.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.CreatedResponse(c, product)
}

// GetProduct retrieves a product
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	product, err := h.service.GetProduct(c.Request.Context(), id)
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, product)
}

// GetProducts retrieves products
func (h *ProductHandler) GetProducts(c *gin.Context) {
	branchID := c.Query("branch_id")
	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// If branch_id not provided, use from context/token user's branch usually?
	// But admin might query any branch.
	// For now require client to send it or default to something?
	// If empty, service fails?
	// User Context has branch_id.
	if branchID == "" {
		if v, exists := c.Get("branch_id"); exists {
			branchID = v.(string)
		}
	}

	products, err := h.service.GetProducts(c.Request.Context(), branchID, status, limit, offset)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, products)
}

// UpdateProduct updates a product
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var req product.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	product, err := h.service.UpdateProduct(c.Request.Context(), id, &req)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, product)
}

// DeleteProduct deletes a product
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteProduct(c.Request.Context(), id); err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.MessageResponse(c, "Product deleted successfully")
}

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
}

// CreateCategory handles product category creation
func (h *ProductHandler) CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	category, err := h.service.CreateCategory(c.Request.Context(), req.Name, req.Code, req.Description)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.CreatedResponse(c, category)
}

// GetCategories retrieves all product categories
func (h *ProductHandler) GetCategories(c *gin.Context) {
	categories, err := h.service.GetCategories(c.Request.Context())
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, categories)
}
