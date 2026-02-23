package handler

import (
	"github.com/devper-gold/gold-shop-api/app/feature/inventory"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
)

type InventoryHandler struct {
	service *inventory.Service
}

func NewInventoryHandler(service *inventory.Service) *InventoryHandler {
	return &InventoryHandler{service: service}
}

// CreateTransfer creates a stock transfer request
func (h *InventoryHandler) CreateTransfer(c *gin.Context) {
	var req inventory.CreateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "INV-400-001", err.Error())
		return
	}

	transfer, err := h.service.CreateTransfer(c.Request.Context(), req)
	if err != nil {
		utils.InternalErrorResponse(c, "INV-500-001", err.Error())
		return
	}

	utils.CreatedResponse(c, transfer)
}

// GetTransfers retrieves stock transfers
func (h *InventoryHandler) GetTransfers(c *gin.Context) {
	branchID := c.Query("branch_id")
	// If branchID not provided, maybe check user's branch?
	// For now, require branch_id or implement generic search in service.

	if branchID == "" {
		// Just for MVP, returns error or assume all if admin?
		// Service currently requires branchID.
		utils.BadRequestResponse(c, "INV-400-002", "branch_id is required")
		return
	}

	transfers, err := h.service.GetTransfers(c.Request.Context(), branchID, "")
	if err != nil {
		utils.InternalErrorResponse(c, "INV-500-002", err.Error())
		return
	}

	utils.SuccessResponse(c, transfers)
}

// GetTransfer retrieves a single transfer
func (h *InventoryHandler) GetTransfer(c *gin.Context) {
	id := c.Param("id")
	transfer, err := h.service.GetTransfer(c.Request.Context(), id)
	if err != nil {
		utils.InternalErrorResponse(c, "INV-500-003", err.Error())
		return
	}

	utils.SuccessResponse(c, transfer)
}

// ApproveTransfer approves a transfer
func (h *InventoryHandler) ApproveTransfer(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID") // From middleware

	if err := h.service.ApproveTransfer(c.Request.Context(), id, userID); err != nil {
		utils.InternalErrorResponse(c, "INV-500-004", err.Error())
		return
	}

	utils.MessageResponse(c, "Transfer approved successfully")
}

// ReceiveTransfer receives a transfer
func (h *InventoryHandler) ReceiveTransfer(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	if err := h.service.ReceiveTransfer(c.Request.Context(), id, userID); err != nil {
		utils.InternalErrorResponse(c, "INV-500-005", err.Error())
		return
	}

	utils.MessageResponse(c, "Transfer received successfully")
}

// CancelTransfer cancels a transfer
func (h *InventoryHandler) CancelTransfer(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.CancelTransfer(c.Request.Context(), id); err != nil {
		utils.InternalErrorResponse(c, "INV-500-006", err.Error())
		return
	}

	utils.MessageResponse(c, "Transfer cancelled successfully")
}
