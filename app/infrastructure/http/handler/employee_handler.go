package handler

import (
	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/feature/employee"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
)

// EmployeeHandler handles employee endpoints
type EmployeeHandler struct {
	service *employee.Service
}

// NewEmployeeHandler creates a new EmployeeHandler
func NewEmployeeHandler(service *employee.Service) *EmployeeHandler {
	return &EmployeeHandler{service: service}
}

// CreateEmployee handles employee creation
func (h *EmployeeHandler) CreateEmployee(c *gin.Context) {
	var req employee.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	createdBy := c.GetString("UserId")
	emp, err := h.service.CreateEmployee(c.Request.Context(), &req, createdBy)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.CreatedResponse(c, emp)
}

// GetEmployees handles getting all employees
func (h *EmployeeHandler) GetEmployees(c *gin.Context) {
	employees, err := h.service.GetEmployees(c.Request.Context())
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, employees)
}

// GetMyEmployee handles getting the current user's employee record
func (h *EmployeeHandler) GetMyEmployee(c *gin.Context) {
	userID := c.GetString("UserId")
	emp, err := h.service.GetMyEmployee(c.Request.Context(), userID)
	if err != nil {
		if err == entity.ErrNotFound {
			utils.NotFoundResponse(c, "Employee not found")
			return
		}
		utils.InternalErrorResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, emp)
}

// GetEmployee handles getting an employee by ID
func (h *EmployeeHandler) GetEmployee(c *gin.Context) {
	id := c.Param("id")
	emp, err := h.service.GetEmployee(c.Request.Context(), id)
	if err != nil {
		if err == entity.ErrNotFound {
			utils.NotFoundResponse(c, "Employee not found")
			return
		}
		utils.InternalErrorResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, emp)
}

// GetEmployeesByBranch handles getting employees by branch ID
func (h *EmployeeHandler) GetEmployeesByBranch(c *gin.Context) {
	branchID := c.Param("branchId")
	employees, err := h.service.GetEmployeesByBranchID(c.Request.Context(), branchID)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, employees)
}

// UpdateEmployee handles employee update
func (h *EmployeeHandler) UpdateEmployee(c *gin.Context) {
	id := c.Param("id")
	var req employee.UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	updatedBy := c.GetString("UserId")
	emp, err := h.service.UpdateEmployee(c.Request.Context(), id, &req, updatedBy)
	if err != nil {
		if err == entity.ErrNotFound {
			utils.NotFoundResponse(c, "Employee not found")
			return
		}
		utils.BadRequestResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, emp)
}

// DeleteEmployee handles employee deletion
func (h *EmployeeHandler) DeleteEmployee(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteEmployee(c.Request.Context(), id); err != nil {
		if err == entity.ErrNotFound {
			utils.NotFoundResponse(c, "Employee not found")
			return
		}
		utils.InternalErrorResponse(c, err.Error())
		return
	}
	utils.MessageResponse(c, "Employee deleted successfully")
}
