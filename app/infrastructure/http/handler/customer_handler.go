package handler

import (
	"strconv"

	"github.com/devper-gold/gold-shop-api/app/feature/customer"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	service *customer.Service
}

func NewCustomerHandler(service *customer.Service) *CustomerHandler {
	return &CustomerHandler{service: service}
}

// GetCustomers retrieves all customers with pagination
func (h *CustomerHandler) GetCustomers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	query := c.Query("q")

	if query != "" {
		customers, err := h.service.SearchCustomers(c.Request.Context(), query, limit)
		if err != nil {
			utils.InternalErrorResponse(c, "CUS-500-001", err.Error())
			return
		}
		utils.SuccessResponse(c, customers)
		return
	}

	customers, err := h.service.GetCustomers(c.Request.Context(), limit, offset)
	if err != nil {
		utils.InternalErrorResponse(c, "CUS-500-002", err.Error())
		return
	}

	utils.SuccessResponse(c, customers)
}

// GetCustomer retrieves a customer by ID
func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	id := c.Param("id")
	customer, err := h.service.GetCustomer(c.Request.Context(), id)
	if err != nil {
		utils.InternalErrorResponse(c, "CUS-500-003", err.Error())
		return
	}
	if customer == nil {
		utils.NotFoundResponse(c, "CUS-404-001", "customer not found")
		return
	}

	utils.SuccessResponse(c, customer)
}

// CreateCustomer creates a new customer
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var req customer.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "CUS-400-001", err.Error())
		return
	}

	newCustomer, err := h.service.CreateCustomer(c.Request.Context(), req)
	if err != nil {
		utils.InternalErrorResponse(c, "CUS-500-004", err.Error())
		return
	}

	utils.CreatedResponse(c, newCustomer)
}

// UpdateCustomer updates an existing customer
func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	id := c.Param("id")
	var req customer.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "CUS-400-002", err.Error())
		return
	}

	updatedCustomer, err := h.service.UpdateCustomer(c.Request.Context(), id, req)
	if err != nil {
		utils.InternalErrorResponse(c, "CUS-500-005", err.Error())
		return
	}

	utils.SuccessResponse(c, updatedCustomer)
}

// DeleteCustomer deletes a customer by ID
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteCustomer(c.Request.Context(), id); err != nil {
		utils.InternalErrorResponse(c, "CUS-500-006", err.Error())
		return
	}

	utils.MessageResponse(c, "Customer deleted successfully")
}

// GetByRFID retrieves a customer by RFID
func (h *CustomerHandler) GetByRFID(c *gin.Context) {
	rfid := c.Param("rfid")
	customer, err := h.service.GetByRFID(c.Request.Context(), rfid)
	if err != nil {
		utils.InternalErrorResponse(c, "CUS-500-007", err.Error())
		return
	}
	if customer == nil {
		utils.NotFoundResponse(c, "CUS-404-002", "customer not found")
		return
	}

	utils.SuccessResponse(c, customer)
}
