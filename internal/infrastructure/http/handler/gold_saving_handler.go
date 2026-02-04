package handler

import (
	"github.com/devper-gold/gold-shop-api/internal/application/gold_saving"
	"github.com/devper-gold/gold-shop-api/internal/domain/entity"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GoldSavingHandler handles gold saving endpoints
type GoldSavingHandler struct {
	goldSavingService *gold_saving.Service
}

// NewGoldSavingHandler creates a new GoldSavingHandler
func NewGoldSavingHandler(goldSavingService *gold_saving.Service) *GoldSavingHandler {
	return &GoldSavingHandler{
		goldSavingService: goldSavingService,
	}
}

// OpenAccountRequest represents open account request
type OpenAccountRequest struct {
	CustomerID    string  `json:"customer_id" binding:"required"`
	SavingType    string  `json:"saving_type" binding:"required"`
	MinDeposit    float64 `json:"min_deposit"`
	MinWithdrawal float64 `json:"min_withdrawal"`
}

// DepositRequest represents deposit request
type DepositRequest struct {
	Amount float64 `json:"amount" binding:"required"`
}

// WithdrawRequest represents withdraw request
type WithdrawRequest struct {
	Amount float64 `json:"amount" binding:"required"`
	AsCash bool    `json:"as_cash"`
}

// List returns gold saving accounts
func (h *GoldSavingHandler) List(c *gin.Context) {
	branchID, _ := primitive.ObjectIDFromHex(c.GetString("branch_id"))

	var status []entity.GoldSavingStatus
	if s := c.Query("status"); s != "" {
		status = []entity.GoldSavingStatus{entity.GoldSavingStatus(s)}
	}

	accounts, err := h.goldSavingService.GetByBranchID(c.Request.Context(), branchID, status)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch accounts")
		return
	}

	utils.SuccessResponse(c, accounts)
}

// Get returns a gold saving account by ID
func (h *GoldSavingHandler) Get(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid account ID")
		return
	}

	account, err := h.goldSavingService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == entity.ErrNotFound {
			utils.NotFoundResponse(c, "Account not found")
			return
		}
		utils.InternalErrorResponse(c, "Failed to fetch account")
		return
	}

	utils.SuccessResponse(c, account)
}

// OpenAccount opens a new gold saving account
func (h *GoldSavingHandler) OpenAccount(c *gin.Context) {
	var req OpenAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	branchID, _ := primitive.ObjectIDFromHex(c.GetString("branch_id"))
	customerID, _ := primitive.ObjectIDFromHex(req.CustomerID)

	account, err := h.goldSavingService.OpenAccount(
		c.Request.Context(),
		branchID,
		customerID,
		entity.GoldSavingType(req.SavingType),
		req.MinDeposit,
		req.MinWithdrawal,
	)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.CreatedResponse(c, account)
}

// Deposit deposits to a gold saving account
func (h *GoldSavingHandler) Deposit(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid account ID")
		return
	}

	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	userID, _ := primitive.ObjectIDFromHex(c.GetString("user_id"))

	account, err := h.goldSavingService.Deposit(c.Request.Context(), id, req.Amount, userID)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, account)
}

// Withdraw withdraws from a gold saving account
func (h *GoldSavingHandler) Withdraw(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid account ID")
		return
	}

	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	userID, _ := primitive.ObjectIDFromHex(c.GetString("user_id"))

	account, err := h.goldSavingService.Withdraw(c.Request.Context(), id, req.Amount, req.AsCash, userID)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, account)
}

// Close closes a gold saving account
func (h *GoldSavingHandler) Close(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid account ID")
		return
	}

	account, err := h.goldSavingService.Close(c.Request.Context(), id)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, account)
}

// GetStatement returns account statement
func (h *GoldSavingHandler) GetStatement(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid account ID")
		return
	}

	statement, err := h.goldSavingService.GetStatement(c.Request.Context(), id)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to generate statement")
		return
	}

	utils.SuccessResponse(c, statement)
}
