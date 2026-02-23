package handler

import (
	"strconv"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/feature/pawn"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PawnHandler handles pawn endpoints
type PawnHandler struct {
	pawnService *pawn.Service
}

// NewPawnHandler creates a new PawnHandler
func NewPawnHandler(pawnService *pawn.Service) *PawnHandler {
	return &PawnHandler{
		pawnService: pawnService,
	}
}

// CreatePawnRequest represents create pawn request
type CreatePawnRequest struct {
	CustomerID   string            `json:"customer_id" binding:"required"`
	Items        []PawnItemRequest `json:"items" binding:"required"`
	Principal    float64           `json:"principal" binding:"required"`
	InterestRate float64           `json:"interest_rate" binding:"required"`
	TermMonths   int               `json:"term_months" binding:"required"`
	Notes        string            `json:"notes"`
}

// PawnItemRequest represents a pawn item in request
type PawnItemRequest struct {
	Description    string   `json:"description" binding:"required"`
	GoldType       string   `json:"gold_type" binding:"required"`
	Weight         float64  `json:"weight" binding:"required"`
	AppraisedValue float64  `json:"appraised_value" binding:"required"`
	Images         []string `json:"images"`
}

// PayInterestRequest represents pay interest request
type PayInterestRequest struct {
	Amount float64 `json:"amount" binding:"required"`
}

// RedeemRequest represents redeem request
type RedeemRequest struct {
	Interest float64 `json:"interest" binding:"required"`
	Discount float64 `json:"discount"`
}

// List returns pawns with filters
func (h *PawnHandler) List(c *gin.Context) {
	branchID, _ := primitive.ObjectIDFromHex(c.Query("branch_id"))
	if branchID.IsZero() {
		branchID, _ = primitive.ObjectIDFromHex(c.GetString("BranchId"))
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var status []entity.PawnStatus
	if s := c.Query("status"); s != "" {
		status = []entity.PawnStatus{entity.PawnStatus(s)}
	}

	pawns, err := h.pawnService.GetByBranchID(c.Request.Context(), branchID, status, limit, offset)
	if err != nil {
		utils.InternalErrorResponse(c, "PWN-500-001", "Failed to fetch pawns")
		return
	}

	utils.SuccessResponse(c, pawns)
}

// Get returns a pawn by ID
func (h *PawnHandler) Get(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "PWN-400-001", "Invalid pawn ID")
		return
	}

	pawnEntity, err := h.pawnService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == entity.ErrNotFound {
			utils.NotFoundResponse(c, "PWN-404-001", "Pawn not found")
			return
		}
		utils.InternalErrorResponse(c, "PWN-500-002", "Failed to fetch pawn")
		return
	}

	utils.SuccessResponse(c, pawnEntity)
}

// Create creates a new pawn
func (h *PawnHandler) Create(c *gin.Context) {
	var req CreatePawnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "PWN-400-002", "Invalid request body")
		return
	}

	branchID, _ := primitive.ObjectIDFromHex(c.GetString("BranchId"))
	userID, _ := primitive.ObjectIDFromHex(c.GetString("UserId"))
	customerID, _ := primitive.ObjectIDFromHex(req.CustomerID)

	input := pawn.CreatePawnInput{
		BranchID:     branchID,
		CustomerID:   customerID,
		UserID:       userID,
		Principal:    req.Principal,
		InterestRate: req.InterestRate,
		TermMonths:   req.TermMonths,
		Notes:        req.Notes,
	}

	for _, item := range req.Items {
		input.Items = append(input.Items, pawn.PawnItemInput{
			Description:    item.Description,
			GoldType:       item.GoldType,
			Weight:         item.Weight,
			AppraisedValue: item.AppraisedValue,
			Images:         item.Images,
		})
	}

	pawnEntity, err := h.pawnService.Create(c.Request.Context(), input)
	if err != nil {
		utils.BadRequestResponse(c, "PWN-400-003", err.Error())
		return
	}

	utils.CreatedResponse(c, pawnEntity)
}

// PayInterest pays interest on a pawn
func (h *PawnHandler) PayInterest(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "PWN-400-004", "Invalid pawn ID")
		return
	}

	var req PayInterestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "PWN-400-005", "Invalid request body")
		return
	}

	userID, _ := primitive.ObjectIDFromHex(c.GetString("UserId"))

	pawnEntity, err := h.pawnService.PayInterest(c.Request.Context(), id, req.Amount, userID)
	if err != nil {
		utils.BadRequestResponse(c, "PWN-400-006", err.Error())
		return
	}

	utils.SuccessResponse(c, pawnEntity)
}

// Redeem redeems a pawn
func (h *PawnHandler) Redeem(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "PWN-400-007", "Invalid pawn ID")
		return
	}

	var req RedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "PWN-400-008", "Invalid request body")
		return
	}

	userID, _ := primitive.ObjectIDFromHex(c.GetString("UserId"))

	pawnEntity, err := h.pawnService.Redeem(c.Request.Context(), id, req.Interest, req.Discount, userID)
	if err != nil {
		utils.BadRequestResponse(c, "PWN-400-009", err.Error())
		return
	}

	utils.SuccessResponse(c, pawnEntity)
}

// Extend extends a pawn term
func (h *PawnHandler) Extend(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "PWN-400-010", "Invalid pawn ID")
		return
	}

	var req struct {
		AdditionalMonths int `json:"additional_months" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "PWN-400-011", "Invalid request body")
		return
	}

	pawnEntity, err := h.pawnService.Extend(c.Request.Context(), id, req.AdditionalMonths)
	if err != nil {
		utils.BadRequestResponse(c, "PWN-400-012", err.Error())
		return
	}

	utils.SuccessResponse(c, pawnEntity)
}

// Forfeit forfeits a pawn
func (h *PawnHandler) Forfeit(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "PWN-400-013", "Invalid pawn ID")
		return
	}

	pawnEntity, err := h.pawnService.Forfeit(c.Request.Context(), id)
	if err != nil {
		utils.BadRequestResponse(c, "PWN-400-014", err.Error())
		return
	}

	utils.SuccessResponse(c, pawnEntity)
}

// GetDueSoon returns pawns due soon
func (h *PawnHandler) GetDueSoon(c *gin.Context) {
	branchID, _ := primitive.ObjectIDFromHex(c.GetString("BranchId"))
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))

	pawns, err := h.pawnService.GetDueSoon(c.Request.Context(), branchID, days)
	if err != nil {
		utils.InternalErrorResponse(c, "PWN-500-003", "Failed to fetch pawns")
		return
	}

	utils.SuccessResponse(c, pawns)
}
