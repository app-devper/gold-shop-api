package branch

import (
	"context"

	"github.com/devper-gold/gold-shop-api/internal/domain/entity"
	"github.com/devper-gold/gold-shop-api/internal/domain/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service handles branch logic
type Service struct {
	branchRepo repository.BranchRepository
}

// NewService creates a new Branch service
func NewService(branchRepo repository.BranchRepository) *Service {
	return &Service{
		branchRepo: branchRepo,
	}
}

// Create creates a new branch
func (s *Service) Create(ctx context.Context, branch *entity.Branch) error {
	return s.branchRepo.Create(ctx, branch)
}

// GetByID retrieves a branch by ID
func (s *Service) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Branch, error) {
	return s.branchRepo.GetByID(ctx, id)
}

// GetByCode retrieves a branch by code
func (s *Service) GetByCode(ctx context.Context, code string) (*entity.Branch, error) {
	return s.branchRepo.GetByCode(ctx, code)
}

// GetAll retrieves all branches
func (s *Service) GetAll(ctx context.Context) ([]*entity.Branch, error) {
	return s.branchRepo.GetAll(ctx)
}

// Update updates a branch
func (s *Service) Update(ctx context.Context, branch *entity.Branch) error {
	return s.branchRepo.Update(ctx, branch)
}

// Delete deletes a branch
func (s *Service) Delete(ctx context.Context, id primitive.ObjectID) error {
	return s.branchRepo.Delete(ctx, id)
}
