package handler

import (
	"strconv"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/feature/expense"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ExpenseHandler struct {
	service *expense.Service
}

func NewExpenseHandler(service *expense.Service) *ExpenseHandler {
	return &ExpenseHandler{service: service}
}

func (h *ExpenseHandler) GetExpenses(c *gin.Context) {
	branchID, _ := primitive.ObjectIDFromHex(c.Query("branch_id"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	expenses, err := h.service.GetExpenses(c.Request.Context(), branchID, limit, offset)
	if err != nil {
		utils.InternalErrorResponse(c, "EXP-500-001", err.Error())
		return
	}
	utils.SuccessResponse(c, expenses)
}

func (h *ExpenseHandler) CreateExpense(c *gin.Context) {
	var req struct {
		CategoryID    string  `json:"category_id" binding:"required"`
		BranchID      string  `json:"branch_id" binding:"required"`
		Amount        float64 `json:"amount" binding:"required"`
		Description   string  `json:"description"`
		ExpenseDate   string  `json:"expense_date"`
		ReceiptNumber string  `json:"receipt_number"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "EXP-400-001", err.Error())
		return
	}

	catOID, _ := primitive.ObjectIDFromHex(req.CategoryID)
	brOID, _ := primitive.ObjectIDFromHex(req.BranchID)
	createdBy, _ := primitive.ObjectIDFromHex(c.GetString("userID"))
	date, _ := time.Parse(time.RFC3339, req.ExpenseDate)
	if req.ExpenseDate == "" {
		date = time.Now()
	}

	expense := entity.NewExpense(brOID, catOID, createdBy, "EXP-"+strconv.FormatInt(time.Now().Unix(), 10), req.Description, req.Amount, date)
	expense.ReceiptNumber = req.ReceiptNumber

	if err := h.service.CreateExpense(c.Request.Context(), expense); err != nil {
		utils.InternalErrorResponse(c, "EXP-500-002", err.Error())
		return
	}

	utils.CreatedResponse(c, expense)
}

func (h *ExpenseHandler) GetCategories(c *gin.Context) {
	categories, err := h.service.GetCategories(c.Request.Context())
	if err != nil {
		utils.InternalErrorResponse(c, "EXP-500-003", err.Error())
		return
	}
	utils.SuccessResponse(c, categories)
}

func (h *ExpenseHandler) CreateCategory(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "EXP-400-002", err.Error())
		return
	}

	category := &entity.ExpenseCategory{
		ID:        primitive.NewObjectID(),
		Code:      req.Code,
		Name:      req.Name,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.service.CreateCategory(c.Request.Context(), category); err != nil {
		utils.InternalErrorResponse(c, "EXP-500-004", err.Error())
		return
	}

	utils.CreatedResponse(c, category)
}
