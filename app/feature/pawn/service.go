package pawn

import (
	"context"
	"errors"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service handles pawn logic
type Service struct {
	pawnRepo   repository.PawnRepository
	branchRepo repository.BranchRepository
}

// NewService creates a new Pawn service
func NewService(pawnRepo repository.PawnRepository, branchRepo repository.BranchRepository) *Service {
	return &Service{
		pawnRepo:   pawnRepo,
		branchRepo: branchRepo,
	}
}

// CreatePawnInput represents input for creating a pawn
type CreatePawnInput struct {
	BranchID     primitive.ObjectID
	CustomerID   primitive.ObjectID
	UserID       primitive.ObjectID
	Items        []PawnItemInput
	Principal    float64
	InterestRate float64
	TermMonths   int
	Notes        string
}

// PawnItemInput represents input for a pawn item
type PawnItemInput struct {
	Description    string
	GoldType       string
	Weight         float64
	AppraisedValue float64
	Images         []string
}

// Create creates a new pawn
func (s *Service) Create(ctx context.Context, input CreatePawnInput) (*entity.Pawn, error) {
	// Get branch for pawn number generation
	branch, err := s.branchRepo.GetByID(ctx, input.BranchID)
	if err != nil || branch == nil {
		return nil, errors.New("branch not found")
	}

	// Generate pawn number
	pawnNumber, err := s.pawnRepo.GeneratePawnNumber(ctx, branch.Code)
	if err != nil {
		return nil, errors.New("failed to generate pawn number")
	}

	// Create pawn
	pawn := entity.NewPawn(
		input.BranchID,
		input.CustomerID,
		input.UserID,
		pawnNumber,
		input.Principal,
		input.InterestRate,
		input.TermMonths,
	)
	pawn.Notes = input.Notes

	// Add items
	for _, itemInput := range input.Items {
		item := entity.PawnItem{
			Description:    itemInput.Description,
			GoldType:       itemInput.GoldType,
			Weight:         itemInput.Weight,
			AppraisedValue: itemInput.AppraisedValue,
			Images:         itemInput.Images,
		}
		pawn.AddItem(item)
	}

	// Save pawn
	if err := s.pawnRepo.Create(ctx, pawn); err != nil {
		return nil, errors.New("failed to create pawn")
	}

	return pawn, nil
}

// GetByID retrieves a pawn by ID
func (s *Service) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Pawn, error) {
	return s.pawnRepo.GetByID(ctx, id)
}

// GetByBranchID retrieves pawns by branch ID
func (s *Service) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.PawnStatus, limit, offset int) ([]*entity.Pawn, error) {
	return s.pawnRepo.GetByBranchID(ctx, branchID, status, limit, offset)
}

// GetDueSoon retrieves pawns due soon
func (s *Service) GetDueSoon(ctx context.Context, branchID primitive.ObjectID, days int) ([]*entity.Pawn, error) {
	return s.pawnRepo.GetDueSoon(ctx, branchID, days)
}

// PayInterest pays interest on a pawn
func (s *Service) PayInterest(ctx context.Context, id primitive.ObjectID, amount float64, userID primitive.ObjectID) (*entity.Pawn, error) {
	pawn, err := s.pawnRepo.GetByID(ctx, id)
	if err != nil || pawn == nil {
		return nil, errors.New("pawn not found")
	}

	if pawn.Status != entity.PawnStatusActive && pawn.Status != entity.PawnStatusExtended {
		return nil, errors.New("pawn is not active")
	}

	// Calculate period
	periodFrom := pawn.StartDate
	if len(pawn.InterestPayments) > 0 {
		periodFrom = pawn.InterestPayments[len(pawn.InterestPayments)-1].PeriodTo
	}
	periodTo := time.Now()

	pawn.PayInterest(amount, periodFrom, periodTo, userID)

	if err := s.pawnRepo.Update(ctx, pawn); err != nil {
		return nil, errors.New("failed to update pawn")
	}

	return pawn, nil
}

// Redeem redeems a pawn
func (s *Service) Redeem(ctx context.Context, id primitive.ObjectID, interest, discount float64, userID primitive.ObjectID) (*entity.Pawn, error) {
	pawn, err := s.pawnRepo.GetByID(ctx, id)
	if err != nil || pawn == nil {
		return nil, errors.New("pawn not found")
	}

	if pawn.Status != entity.PawnStatusActive && pawn.Status != entity.PawnStatusExtended {
		return nil, errors.New("pawn is not active")
	}

	pawn.Redeem(interest, discount, userID)

	if err := s.pawnRepo.Update(ctx, pawn); err != nil {
		return nil, errors.New("failed to update pawn")
	}

	return pawn, nil
}

// Extend extends a pawn term
func (s *Service) Extend(ctx context.Context, id primitive.ObjectID, additionalMonths int) (*entity.Pawn, error) {
	pawn, err := s.pawnRepo.GetByID(ctx, id)
	if err != nil || pawn == nil {
		return nil, errors.New("pawn not found")
	}

	if pawn.Status != entity.PawnStatusActive {
		return nil, errors.New("pawn is not active")
	}

	pawn.Extend(additionalMonths)

	if err := s.pawnRepo.Update(ctx, pawn); err != nil {
		return nil, errors.New("failed to update pawn")
	}

	return pawn, nil
}

// Forfeit forfeits a pawn
func (s *Service) Forfeit(ctx context.Context, id primitive.ObjectID) (*entity.Pawn, error) {
	pawn, err := s.pawnRepo.GetByID(ctx, id)
	if err != nil || pawn == nil {
		return nil, errors.New("pawn not found")
	}

	if pawn.Status != entity.PawnStatusActive && pawn.Status != entity.PawnStatusExtended {
		return nil, errors.New("pawn is not active")
	}

	pawn.Forfeit()

	if err := s.pawnRepo.Update(ctx, pawn); err != nil {
		return nil, errors.New("failed to update pawn")
	}

	return pawn, nil
}
