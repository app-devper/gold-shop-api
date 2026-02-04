package pawn

import (
	"context"
	"testing"
	"time"

	"github.com/devper-gold/gold-shop-api/internal/domain/entity"
	"github.com/devper-gold/gold-shop-api/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCreatePawn(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		branchID := primitive.NewObjectID()
		customerID := primitive.NewObjectID()
		userID := primitive.NewObjectID()

		mockPawnRepo := new(testutils.MockPawnRepository)
		mockBranchRepo := new(testutils.MockBranchRepository)
		service := NewService(mockPawnRepo, mockBranchRepo)

		branch := &entity.Branch{ID: branchID, Code: "B001", Name: "Test Branch"}

		mockBranchRepo.On("GetByID", ctx, branchID).Return(branch, nil)
		mockPawnRepo.On("GeneratePawnNumber", ctx, "B001").Return("P20240001", nil)
		mockPawnRepo.On("Create", ctx, mock.AnythingOfType("*entity.Pawn")).Return(nil)

		input := CreatePawnInput{
			BranchID:     branchID,
			CustomerID:   customerID,
			UserID:       userID,
			Principal:    10000,
			InterestRate: 2.5,
			TermMonths:   4,
			Items: []PawnItemInput{
				{
					Description:    "Gold Ring",
					GoldType:       "96.5%",
					Weight:         15.16,
					AppraisedValue: 12000,
				},
			},
		}

		pawn, err := service.Create(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, pawn)
		assert.Equal(t, "P20240001", pawn.PawnNumber)
		assert.Equal(t, 10000.0, pawn.Principal)
		assert.Equal(t, 1, len(pawn.Items))

		mockBranchRepo.AssertExpectations(t)
		mockPawnRepo.AssertExpectations(t)
	})

	t.Run("BranchNotFound", func(t *testing.T) {
		branchID := primitive.NewObjectID()
		mockBranchRepo := new(testutils.MockBranchRepository)
		service := NewService(nil, mockBranchRepo)

		mockBranchRepo.On("GetByID", ctx, branchID).Return(nil, nil).Once()

		input := CreatePawnInput{BranchID: branchID}
		pawn, err := service.Create(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, pawn)
		assert.Equal(t, "branch not found", err.Error())
	})
}

func TestPayInterest(t *testing.T) {
	ctx := context.Background()
	pawnID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	t.Run("Success", func(t *testing.T) {
		mockPawnRepo := new(testutils.MockPawnRepository)
		service := NewService(mockPawnRepo, nil)

		pawn := &entity.Pawn{
			ID:        pawnID,
			Status:    entity.PawnStatusActive,
			StartDate: time.Now().AddDate(0, -1, 0),
		}

		mockPawnRepo.On("GetByID", ctx, pawnID).Return(pawn, nil)
		mockPawnRepo.On("Update", ctx, mock.AnythingOfType("*entity.Pawn")).Return(nil)

		updatedPawn, err := service.PayInterest(ctx, pawnID, 250, userID)

		assert.NoError(t, err)
		assert.NotNil(t, updatedPawn)
		assert.Equal(t, 1, len(updatedPawn.InterestPayments))
		assert.Equal(t, 250.0, updatedPawn.InterestPayments[0].Amount)

		mockPawnRepo.AssertExpectations(t)
	})

	t.Run("PawnNotFound", func(t *testing.T) {
		mockPawnRepo := new(testutils.MockPawnRepository)
		service := NewService(mockPawnRepo, nil)

		mockPawnRepo.On("GetByID", ctx, pawnID).Return(nil, nil).Once()

		updatedPawn, err := service.PayInterest(ctx, pawnID, 250, userID)

		assert.Error(t, err)
		assert.Nil(t, updatedPawn)
		assert.Equal(t, "pawn not found", err.Error())
	})
}
