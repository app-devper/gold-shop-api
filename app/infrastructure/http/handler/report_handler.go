package handler

import (
	"strconv"

	"github.com/devper-gold/gold-shop-api/app/feature/report"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReportHandler handles report endpoints
type ReportHandler struct {
	reportService *report.Service
}

// NewReportHandler creates a new ReportHandler
func NewReportHandler(reportService *report.Service) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
	}
}

// GetDashboardData returns dashboard statistics
func (h *ReportHandler) GetDashboardData(c *gin.Context) {
	branchID, _ := primitive.ObjectIDFromHex(c.GetString("BranchId"))
	if branchID.IsZero() {
		// If not in context (admin cases), try query param
		branchID, _ = primitive.ObjectIDFromHex(c.Query("branch_id"))
	}

	if branchID.IsZero() {
		utils.BadRequestResponse(c, "REP-400-001", "Branch ID is required")
		return
	}

	data, err := h.reportService.GetDashboardData(c.Request.Context(), branchID)
	if err != nil {
		utils.InternalErrorResponse(c, "REP-500-001", err.Error())
		return
	}

	utils.SuccessResponse(c, data)
}

// GetProfitLossReport returns a P&L report
func (h *ReportHandler) GetProfitLossReport(c *gin.Context) {
	branchID, _ := primitive.ObjectIDFromHex(c.GetString("BranchId"))
	if branchID.IsZero() {
		branchID, _ = primitive.ObjectIDFromHex(c.Query("branch_id"))
	}

	if branchID.IsZero() {
		utils.BadRequestResponse(c, "REP-400-002", "Branch ID is required")
		return
	}

	from := c.Query("from")
	to := c.Query("to")

	if from == "" || to == "" {
		utils.BadRequestResponse(c, "REP-400-003", "Date range (from, to) is required")
		return
	}

	report, err := h.reportService.GetProfitLossReport(c.Request.Context(), branchID, from, to)
	if err != nil {
		utils.InternalErrorResponse(c, "REP-500-002", err.Error())
		return
	}

	utils.SuccessResponse(c, report)
}

// GetMultiBranchProfitLossReport returns P&L reports for all branches
func (h *ReportHandler) GetMultiBranchProfitLossReport(c *gin.Context) {
	// Note: Role check should be done in router
	from := c.Query("from")
	to := c.Query("to")

	if from == "" || to == "" {
		utils.BadRequestResponse(c, "REP-400-004", "Date range (from, to) is required")
		return
	}

	reports, err := h.reportService.GetMultiBranchProfitLossReport(c.Request.Context(), from, to)
	if err != nil {
		utils.InternalErrorResponse(c, "REP-500-003", err.Error())
		return
	}

	utils.SuccessResponse(c, reports)
}

// GetTopProducts returns top selling products
func (h *ReportHandler) GetTopProducts(c *gin.Context) {
	branchID, _ := primitive.ObjectIDFromHex(c.GetString("BranchId"))
	if branchID.IsZero() {
		branchID, _ = primitive.ObjectIDFromHex(c.Query("branch_id"))
	}

	from := c.Query("from")
	to := c.Query("to")
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)

	if from == "" || to == "" {
		utils.BadRequestResponse(c, "REP-400-005", "Date range (from, to) is required")
		return
	}

	data, err := h.reportService.GetTopSellingProducts(c.Request.Context(), branchID, from, to, limit)
	if err != nil {
		utils.InternalErrorResponse(c, "REP-500-004", err.Error())
		return
	}

	utils.SuccessResponse(c, data)
}

// GetEmployeePerformance returns employee sales performance
func (h *ReportHandler) GetEmployeePerformance(c *gin.Context) {
	branchID, _ := primitive.ObjectIDFromHex(c.GetString("BranchId"))
	if branchID.IsZero() {
		branchID, _ = primitive.ObjectIDFromHex(c.Query("branch_id"))
	}

	from := c.Query("from")
	to := c.Query("to")

	if from == "" || to == "" {
		utils.BadRequestResponse(c, "REP-400-006", "Date range (from, to) is required")
		return
	}

	data, err := h.reportService.GetEmployeePerformance(c.Request.Context(), branchID, from, to)
	if err != nil {
		utils.InternalErrorResponse(c, "REP-500-005", err.Error())
		return
	}

	utils.SuccessResponse(c, data)
}

// GetSalesTrends returns daily sales trends
func (h *ReportHandler) GetSalesTrends(c *gin.Context) {
	branchID, _ := primitive.ObjectIDFromHex(c.GetString("BranchId"))
	if branchID.IsZero() {
		branchID, _ = primitive.ObjectIDFromHex(c.Query("branch_id"))
	}

	from := c.Query("from")
	to := c.Query("to")

	if from == "" || to == "" {
		utils.BadRequestResponse(c, "REP-400-007", "Date range (from, to) is required")
		return
	}

	data, err := h.reportService.GetSalesTrends(c.Request.Context(), branchID, from, to)
	if err != nil {
		utils.InternalErrorResponse(c, "REP-500-006", err.Error())
		return
	}

	utils.SuccessResponse(c, data)
}
