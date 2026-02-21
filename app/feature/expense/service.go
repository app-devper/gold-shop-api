package expense

import (
	"context"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	expenseRepo repository.ExpenseRepository
	catRepo     repository.ExpenseCategoryRepository
}

func NewService(expenseRepo repository.ExpenseRepository, catRepo repository.ExpenseCategoryRepository) *Service {
	return &Service{
		expenseRepo: expenseRepo,
		catRepo:     catRepo,
	}
}

func (s *Service) GetExpenses(ctx context.Context, branchID primitive.ObjectID, limit, offset int) ([]*entity.Expense, error) {
	return s.expenseRepo.GetByBranchID(ctx, branchID, nil, limit, offset)
}

func (s *Service) CreateExpense(ctx context.Context, expense *entity.Expense) error {
	return s.expenseRepo.Create(ctx, expense)
}

func (s *Service) ApproveExpense(ctx context.Context, id, approvedBy primitive.ObjectID) error {
	expense, err := s.expenseRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	expense.Approve(approvedBy)
	return s.expenseRepo.Update(ctx, expense)
}

func (s *Service) GetCategories(ctx context.Context) ([]*entity.ExpenseCategory, error) {
	return s.catRepo.GetAll(ctx)
}

func (s *Service) CreateCategory(ctx context.Context, category *entity.ExpenseCategory) error {
	return s.catRepo.Create(ctx, category)
}
