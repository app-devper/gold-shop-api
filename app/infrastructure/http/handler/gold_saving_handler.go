package handler

import (
	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/feature/gold_saving"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GoldSavingHandler handles gold-saving HTTP endpoints.
type GoldSavingHandler struct {
	goldSavingService *gold_saving.Service
}

// NewGoldSavingHandler creates a new GoldSavingHandler.
func NewGoldSavingHandler(svc *gold_saving.Service) *GoldSavingHandler {
	return &GoldSavingHandler{goldSavingService: svc}
}

// OpenAccountRequest is the unified-account creation payload.
// Min thresholds are optional (default 0 = disabled).
type OpenAccountRequest struct {
	CustomerID          string  `json:"customer_id" binding:"required"`
	MinDepositCash      float64 `json:"min_deposit_cash"`
	MinDepositPhysical  float64 `json:"min_deposit_physical"`
	MinWithdrawCash     float64 `json:"min_withdraw_cash"`
	MinWithdrawPhysical float64 `json:"min_withdraw_physical"`
}

// TransactionRequest is shared by deposit and withdraw.
// Mode discriminates the unit of `amount`:
//   - cash     → amount is baht
//   - physical → amount is grams
type TransactionRequest struct {
	Mode   string  `json:"mode" binding:"required,oneof=cash physical"`
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

// AdjustRequest is the admin manual correction payload.
// WeightDelta is signed (grams); Note must explain the reason.
type AdjustRequest struct {
	WeightDelta float64 `json:"weight_delta" binding:"required"`
	Note        string  `json:"note" binding:"required"`
}

// List returns all accounts in the operator's branch (filterable by status).
func (h *GoldSavingHandler) List(c *gin.Context) {
	branchID, _ := primitive.ObjectIDFromHex(c.GetString("BranchId"))

	var status []entity.GoldSavingStatus
	if s := c.Query("status"); s != "" {
		status = []entity.GoldSavingStatus{entity.GoldSavingStatus(s)}
	}
	accounts, err := h.goldSavingService.GetByBranchID(c.Request.Context(), branchID, status)
	if err != nil {
		utils.InternalErrorResponse(c, "GLS-500-001", "Failed to fetch accounts")
		return
	}
	utils.SuccessResponse(c, accounts)
}

// Get returns a single account by ID.
func (h *GoldSavingHandler) Get(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "GLS-400-001", "Invalid account ID")
		return
	}
	account, err := h.goldSavingService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == entity.ErrNotFound {
			utils.NotFoundResponse(c, "GLS-404-001", "Account not found")
			return
		}
		utils.InternalErrorResponse(c, "GLS-500-002", "Failed to fetch account")
		return
	}
	utils.SuccessResponse(c, account)
}

// OpenAccount creates a new account in the operator's branch.
func (h *GoldSavingHandler) OpenAccount(c *gin.Context) {
	var req OpenAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "GLS-400-002", "Invalid request body")
		return
	}
	branchID, _ := primitive.ObjectIDFromHex(c.GetString("BranchId"))
	customerID, err := primitive.ObjectIDFromHex(req.CustomerID)
	if err != nil {
		utils.BadRequestResponse(c, "GLS-400-003", "Invalid customer ID")
		return
	}
	account, err := h.goldSavingService.Open(c.Request.Context(), gold_saving.OpenInput{
		BranchID:            branchID,
		CustomerID:          customerID,
		MinDepositCash:      req.MinDepositCash,
		MinDepositPhysical:  req.MinDepositPhysical,
		MinWithdrawCash:     req.MinWithdrawCash,
		MinWithdrawPhysical: req.MinWithdrawPhysical,
	})
	if err != nil {
		utils.BadRequestResponse(c, "GLS-400-004", err.Error())
		return
	}
	utils.CreatedResponse(c, account)
}

// Deposit dispatches by request.Mode to either cash or physical deposit.
func (h *GoldSavingHandler) Deposit(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "GLS-400-005", "Invalid account ID")
		return
	}
	var req TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "GLS-400-006", "Invalid request body")
		return
	}
	userID, _ := primitive.ObjectIDFromHex(c.GetString("UserId"))

	var account *entity.GoldSaving
	switch entity.TxMode(req.Mode) {
	case entity.TxModeCash:
		account, err = h.goldSavingService.DepositCash(c.Request.Context(), id, req.Amount, userID)
	case entity.TxModePhysical:
		account, err = h.goldSavingService.DepositGold(c.Request.Context(), id, req.Amount, userID)
	default:
		utils.BadRequestResponse(c, "GLS-400-007", "mode must be cash or physical")
		return
	}
	if err != nil {
		utils.BadRequestResponse(c, "GLS-400-008", err.Error())
		return
	}
	utils.SuccessResponse(c, account)
}

// Withdraw dispatches by request.Mode to either cash or physical withdrawal.
func (h *GoldSavingHandler) Withdraw(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "GLS-400-009", "Invalid account ID")
		return
	}
	var req TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "GLS-400-010", "Invalid request body")
		return
	}
	userID, _ := primitive.ObjectIDFromHex(c.GetString("UserId"))

	var account *entity.GoldSaving
	switch entity.TxMode(req.Mode) {
	case entity.TxModeCash:
		account, err = h.goldSavingService.WithdrawCash(c.Request.Context(), id, req.Amount, userID)
	case entity.TxModePhysical:
		account, err = h.goldSavingService.WithdrawGold(c.Request.Context(), id, req.Amount, userID)
	default:
		utils.BadRequestResponse(c, "GLS-400-011", "mode must be cash or physical")
		return
	}
	if err != nil {
		utils.BadRequestResponse(c, "GLS-400-012", err.Error())
		return
	}
	utils.SuccessResponse(c, account)
}

// Adjust applies an admin manual correction. Route is ADMIN-gated.
func (h *GoldSavingHandler) Adjust(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "GLS-400-013", "Invalid account ID")
		return
	}
	var req AdjustRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "GLS-400-014", "Invalid request body")
		return
	}
	userID, _ := primitive.ObjectIDFromHex(c.GetString("UserId"))

	account, err := h.goldSavingService.Adjust(c.Request.Context(), gold_saving.AdjustInput{
		AccountID:   id,
		WeightDelta: req.WeightDelta,
		Note:        req.Note,
		By:          userID,
	})
	if err != nil {
		utils.BadRequestResponse(c, "GLS-400-015", err.Error())
		return
	}
	utils.SuccessResponse(c, account)
}

// Close closes an account (balance must be effectively zero).
func (h *GoldSavingHandler) Close(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "GLS-400-016", "Invalid account ID")
		return
	}
	account, err := h.goldSavingService.Close(c.Request.Context(), id)
	if err != nil {
		utils.BadRequestResponse(c, "GLS-400-017", err.Error())
		return
	}
	utils.SuccessResponse(c, account)
}

// GetStatement returns the mark-to-market statement.
func (h *GoldSavingHandler) GetStatement(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "GLS-400-018", "Invalid account ID")
		return
	}
	statement, err := h.goldSavingService.GetStatement(c.Request.Context(), id)
	if err != nil {
		utils.BadRequestResponse(c, "GLS-400-019", err.Error())
		return
	}
	utils.SuccessResponse(c, statement)
}
