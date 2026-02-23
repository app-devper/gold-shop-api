package handler

import (
	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/feature/branch"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BranchHandler handles branch endpoints
type BranchHandler struct {
	branchService *branch.Service
}

// NewBranchHandler creates a new BranchHandler
func NewBranchHandler(branchService *branch.Service) *BranchHandler {
	return &BranchHandler{
		branchService: branchService,
	}
}

// CreateBranchRequest represents create branch request
type CreateBranchRequest struct {
	Code    string `json:"code" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

// UpdateBranchRequest represents update branch request
type UpdateBranchRequest struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	IsActive *bool  `json:"is_active"`
}

// List returns all branches
func (h *BranchHandler) List(c *gin.Context) {
	branches, err := h.branchService.GetAll(c.Request.Context())
	if err != nil {
		utils.InternalErrorResponse(c, "BRN-500-001", "Failed to fetch branches")
		return
	}

	utils.SuccessResponse(c, branches)
}

// Get returns a branch by ID
func (h *BranchHandler) Get(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "BRN-400-001", "Invalid branch ID")
		return
	}

	branchEntity, err := h.branchService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == entity.ErrNotFound {
			utils.NotFoundResponse(c, "BRN-404-001", "Branch not found")
			return
		}
		utils.InternalErrorResponse(c, "BRN-500-002", "Failed to fetch branch")
		return
	}

	utils.SuccessResponse(c, branchEntity)
}

// Create creates a new branch
func (h *BranchHandler) Create(c *gin.Context) {
	var req CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "BRN-400-002", "Invalid request body")
		return
	}

	branchEntity := entity.NewBranch(req.Code, req.Name, req.Address, req.Phone)
	if err := h.branchService.Create(c.Request.Context(), branchEntity); err != nil {
		if err == entity.ErrDuplicateKey {
			utils.BadRequestResponse(c, "BRN-400-003", "Branch code already exists")
			return
		}
		utils.InternalErrorResponse(c, "BRN-500-003", "Failed to create branch")
		return
	}

	utils.CreatedResponse(c, branchEntity)
}

// Update updates a branch
func (h *BranchHandler) Update(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "BRN-400-004", "Invalid branch ID")
		return
	}

	var req UpdateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "BRN-400-005", "Invalid request body")
		return
	}

	branchEntity, err := h.branchService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == entity.ErrNotFound {
			utils.NotFoundResponse(c, "BRN-404-002", "Branch not found")
			return
		}
		utils.InternalErrorResponse(c, "BRN-500-004", "Failed to fetch branch")
		return
	}

	// Update fields
	if req.Name != "" {
		branchEntity.Name = req.Name
	}
	if req.Address != "" {
		branchEntity.Address = req.Address
	}
	if req.Phone != "" {
		branchEntity.Phone = req.Phone
	}
	if req.IsActive != nil {
		branchEntity.IsActive = *req.IsActive
	}

	if err := h.branchService.Update(c.Request.Context(), branchEntity); err != nil {
		utils.InternalErrorResponse(c, "BRN-500-005", "Failed to update branch")
		return
	}

	utils.SuccessResponse(c, branchEntity)
}

// Delete deletes a branch
func (h *BranchHandler) Delete(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "BRN-400-006", "Invalid branch ID")
		return
	}

	if err := h.branchService.Delete(c.Request.Context(), id); err != nil {
		if err == entity.ErrNotFound {
			utils.NotFoundResponse(c, "BRN-404-003", "Branch not found")
			return
		}
		utils.InternalErrorResponse(c, "BRN-500-006", "Failed to delete branch")
		return
	}

	utils.MessageResponse(c, "Branch deleted successfully")
}
