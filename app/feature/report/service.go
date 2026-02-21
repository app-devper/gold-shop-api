package report

import (
	"context"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service handles reporting logic
type Service struct {
	saleRepo       repository.SaleRepository
	pawnRepo       repository.PawnRepository
	expenseRepo    repository.ExpenseRepository
	goldSavingRepo repository.GoldSavingRepository
	branchRepo     repository.BranchRepository
}

// NewService creates a new Report service
func NewService(
	saleRepo repository.SaleRepository,
	pawnRepo repository.PawnRepository,
	expenseRepo repository.ExpenseRepository,
	goldSavingRepo repository.GoldSavingRepository,
	branchRepo repository.BranchRepository,
) *Service {
	return &Service{
		saleRepo:       saleRepo,
		pawnRepo:       pawnRepo,
		expenseRepo:    expenseRepo,
		goldSavingRepo: goldSavingRepo,
		branchRepo:     branchRepo,
	}
}

// DashboardData represents dashboard statistics
type DashboardData struct {
	TotalSalesToday  float64 `json:"total_sales_today"`
	TransactionCount int64   `json:"transaction_count"`
	ActivePawns      int64   `json:"active_pawns"`
	DueSoonPawns     int     `json:"due_soon_pawns"`
	GoldSavingTotal  float64 `json:"gold_saving_total"`
}

// GetDashboardData retrieves dashboard stats for a branch
func (s *Service) GetDashboardData(ctx context.Context, branchID primitive.ObjectID) (*DashboardData, error) {
	today := time.Now().Format("2006-01-02")

	totalSales, _ := s.saleRepo.SumByBranchAndDateRange(ctx, branchID, today, today)
	countSales, _ := s.saleRepo.CountByBranchAndDateRange(ctx, branchID, today, today)
	activePawns, _ := s.pawnRepo.CountByStatus(ctx, branchID, entity.PawnStatusActive)
	dueSoonPawnsEntities, _ := s.pawnRepo.GetDueSoon(ctx, branchID, 7) // Due in 7 days
	goldSavingTotal, _ := s.goldSavingRepo.SumBalanceByBranch(ctx, branchID)

	return &DashboardData{
		TotalSalesToday:  totalSales,
		TransactionCount: countSales,
		ActivePawns:      activePawns,
		DueSoonPawns:     len(dueSoonPawnsEntities),
		GoldSavingTotal:  goldSavingTotal,
	}, nil
}

// ProfitLossReport represents a profit and loss report
type ProfitLossReport struct {
	BranchID            primitive.ObjectID `json:"branch_id"`
	BranchName          string             `json:"branch_name"`
	PeriodFrom          string             `json:"period_from"`
	PeriodTo            string             `json:"period_to"`
	GoldSaleRevenue     float64            `json:"gold_sale_revenue"`
	PawnInterestRevenue float64            `json:"pawn_interest_revenue"`
	TotalRevenue        float64            `json:"total_revenue"`
	CostOfGoldSold      float64            `json:"cost_of_gold_sold"`
	GrossProfit         float64            `json:"gross_profit"`
	TotalExpenses       float64            `json:"total_expenses"`
	NetProfit           float64            `json:"net_profit"`
}

// GetProfitLossReport generates a P&L report for a branch and period
func (s *Service) GetProfitLossReport(ctx context.Context, branchID primitive.ObjectID, from, to string) (*ProfitLossReport, error) {
	saleRevenue, _ := s.saleRepo.SumByBranchAndDateRange(ctx, branchID, from, to)
	interestRevenue, _ := s.pawnRepo.SumInterestByBranchAndDateRange(ctx, branchID, from, to)
	costOfGold, _ := s.saleRepo.SumCostByBranchAndDateRange(ctx, branchID, from, to)
	totalExpenses, _ := s.expenseRepo.SumByBranchAndDateRange(ctx, branchID, from, to)

	branch, _ := s.branchRepo.GetByID(ctx, branchID)
	branchName := ""
	if branch != nil {
		branchName = branch.Name
	}

	totalRevenue := saleRevenue + interestRevenue
	grossProfit := totalRevenue - costOfGold
	netProfit := grossProfit - totalExpenses

	return &ProfitLossReport{
		BranchID:            branchID,
		BranchName:          branchName,
		PeriodFrom:          from,
		PeriodTo:            to,
		GoldSaleRevenue:     saleRevenue,
		PawnInterestRevenue: interestRevenue,
		TotalRevenue:        totalRevenue,
		CostOfGoldSold:      costOfGold,
		GrossProfit:         grossProfit,
		TotalExpenses:       totalExpenses,
		NetProfit:           netProfit,
	}, nil
}

// GetMultiBranchProfitLossReport generates P&L reports for all branches
func (s *Service) GetMultiBranchProfitLossReport(ctx context.Context, from, to string) ([]*ProfitLossReport, error) {
	branches, err := s.branchRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var reports []*ProfitLossReport
	for _, b := range branches {
		report, _ := s.GetProfitLossReport(ctx, b.ID, from, to)
		if report != nil {
			reports = append(reports, report)
		}
	}

	return reports, nil
}

// GetTopSellingProducts retrieves top selling products
func (s *Service) GetTopSellingProducts(ctx context.Context, branchID primitive.ObjectID, from, to string, limit int) ([]repository.TopProduct, error) {
	return s.saleRepo.GetTopSellingProducts(ctx, branchID, from, to, limit)
}

// GetEmployeePerformance retrieves employee sales performance
func (s *Service) GetEmployeePerformance(ctx context.Context, branchID primitive.ObjectID, from, to string) ([]repository.EmployeePerformance, error) {
	return s.saleRepo.GetEmployeePerformance(ctx, branchID, from, to)
}

// GetSalesTrends retrieves daily sales trends
func (s *Service) GetSalesTrends(ctx context.Context, branchID primitive.ObjectID, from, to string) ([]repository.SalesTrend, error) {
	return s.saleRepo.GetSalesTrends(ctx, branchID, from, to)
}
