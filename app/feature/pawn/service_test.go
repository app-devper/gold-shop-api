package pawn

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/testutils"
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
		service := NewService(mockPawnRepo, mockBranchRepo, nil)

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
		service := NewService(nil, mockBranchRepo, nil)

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
		service := NewService(mockPawnRepo, nil, nil)

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
		service := NewService(mockPawnRepo, nil, nil)

		mockPawnRepo.On("GetByID", ctx, pawnID).Return(nil, nil).Once()

		updatedPawn, err := service.PayInterest(ctx, pawnID, 250, userID)

		assert.Error(t, err)
		assert.Nil(t, updatedPawn)
		assert.Equal(t, "pawn not found", err.Error())
	})
}

func TestGetByBranchIDAttachesCustomerNames(t *testing.T) {
	ctx := context.Background()
	branchID := primitive.NewObjectID()
	customerA := primitive.NewObjectID()
	customerB := primitive.NewObjectID()

	t.Run("Success", func(t *testing.T) {
		mockPawnRepo := new(testutils.MockPawnRepository)
		mockCustomerRepo := new(testutils.MockCustomerRepository)
		service := NewService(mockPawnRepo, nil, mockCustomerRepo)

		pawns := []*entity.Pawn{
			{ID: primitive.NewObjectID(), CustomerID: customerA},
			{ID: primitive.NewObjectID(), CustomerID: customerB},
			{ID: primitive.NewObjectID(), CustomerID: customerA},
		}
		mockPawnRepo.On("GetByBranchID", ctx, branchID, []entity.PawnStatus(nil), 20, 0).Return(pawns, nil)
		mockCustomerRepo.On("GetNamesByIDs", ctx, mock.MatchedBy(func(ids []primitive.ObjectID) bool {
			return len(ids) == 2
		})).Return(map[primitive.ObjectID]string{
			customerA: "สมชาย ใจดี",
			customerB: "สมหญิง รักทอง",
		}, nil)

		result, err := service.GetByBranchID(ctx, branchID, nil, 20, 0)

		assert.NoError(t, err)
		assert.Equal(t, "สมชาย ใจดี", result[0].CustomerName)
		assert.Equal(t, "สมหญิง รักทอง", result[1].CustomerName)
		assert.Equal(t, "สมชาย ใจดี", result[2].CustomerName)
		mockPawnRepo.AssertExpectations(t)
		mockCustomerRepo.AssertExpectations(t)
	})

	t.Run("EmptyList", func(t *testing.T) {
		mockPawnRepo := new(testutils.MockPawnRepository)
		mockCustomerRepo := new(testutils.MockCustomerRepository)
		service := NewService(mockPawnRepo, nil, mockCustomerRepo)

		mockPawnRepo.On("GetByBranchID", ctx, branchID, []entity.PawnStatus(nil), 20, 0).Return([]*entity.Pawn{}, nil)

		result, err := service.GetByBranchID(ctx, branchID, nil, 20, 0)

		assert.NoError(t, err)
		assert.Empty(t, result)
		mockCustomerRepo.AssertNotCalled(t, "GetNamesByIDs")
	})

	t.Run("NameLookupError", func(t *testing.T) {
		mockPawnRepo := new(testutils.MockPawnRepository)
		mockCustomerRepo := new(testutils.MockCustomerRepository)
		service := NewService(mockPawnRepo, nil, mockCustomerRepo)

		pawns := []*entity.Pawn{{ID: primitive.NewObjectID(), CustomerID: customerA}}
		mockPawnRepo.On("GetByBranchID", ctx, branchID, []entity.PawnStatus(nil), 20, 0).Return(pawns, nil)
		mockCustomerRepo.On("GetNamesByIDs", ctx, mock.Anything).Return(nil, errors.New("db down"))

		result, err := service.GetByBranchID(ctx, branchID, nil, 20, 0)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestGetDueSoonAttachesCustomerNames(t *testing.T) {
	ctx := context.Background()
	branchID := primitive.NewObjectID()
	customerID := primitive.NewObjectID()

	mockPawnRepo := new(testutils.MockPawnRepository)
	mockCustomerRepo := new(testutils.MockCustomerRepository)
	service := NewService(mockPawnRepo, nil, mockCustomerRepo)

	pawns := []*entity.Pawn{{ID: primitive.NewObjectID(), CustomerID: customerID}}
	mockPawnRepo.On("GetDueSoon", ctx, branchID, 7).Return(pawns, nil)
	mockCustomerRepo.On("GetNamesByIDs", ctx, []primitive.ObjectID{customerID}).Return(map[primitive.ObjectID]string{
		customerID: "สมปอง มั่งมี",
	}, nil)

	result, err := service.GetDueSoon(ctx, branchID, 7)

	assert.NoError(t, err)
	assert.Equal(t, "สมปอง มั่งมี", result[0].CustomerName)
}

func TestGetByIDAttachesCustomerName(t *testing.T) {
	ctx := context.Background()
	pawnID := primitive.NewObjectID()
	customerID := primitive.NewObjectID()

	t.Run("Success", func(t *testing.T) {
		mockPawnRepo := new(testutils.MockPawnRepository)
		mockCustomerRepo := new(testutils.MockCustomerRepository)
		service := NewService(mockPawnRepo, nil, mockCustomerRepo)

		mockPawnRepo.On("GetByID", ctx, pawnID).Return(&entity.Pawn{ID: pawnID, CustomerID: customerID}, nil)
		mockCustomerRepo.On("GetNamesByIDs", ctx, []primitive.ObjectID{customerID}).Return(map[primitive.ObjectID]string{
			customerID: "สมศรี ศรีทอง",
		}, nil)

		result, err := service.GetByID(ctx, pawnID)

		assert.NoError(t, err)
		assert.Equal(t, "สมศรี ศรีทอง", result.CustomerName)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockPawnRepo := new(testutils.MockPawnRepository)
		mockCustomerRepo := new(testutils.MockCustomerRepository)
		service := NewService(mockPawnRepo, nil, mockCustomerRepo)

		mockPawnRepo.On("GetByID", ctx, pawnID).Return(nil, nil)

		result, err := service.GetByID(ctx, pawnID)

		assert.NoError(t, err)
		assert.Nil(t, result)
		mockCustomerRepo.AssertNotCalled(t, "GetNamesByIDs")
	})
}

func TestSearchMatchesPawnNumberAndCustomerName(t *testing.T) {
	ctx := context.Background()
	branchID := primitive.NewObjectID()
	customerID := primitive.NewObjectID()

	t.Run("Success", func(t *testing.T) {
		mockPawnRepo := new(testutils.MockPawnRepository)
		mockCustomerRepo := new(testutils.MockCustomerRepository)
		service := NewService(mockPawnRepo, nil, mockCustomerRepo)

		mockCustomerRepo.On("Search", ctx, "สมชาย", 5).Return([]*entity.Customer{
			{ID: customerID, FullName: "สมชาย ใจดี"},
		}, nil)
		pawns := []*entity.Pawn{{ID: primitive.NewObjectID(), CustomerID: customerID}}
		mockPawnRepo.On("Search", ctx, branchID, "สมชาย", []primitive.ObjectID{customerID}, 5).Return(pawns, nil)
		mockCustomerRepo.On("GetNamesByIDs", ctx, []primitive.ObjectID{customerID}).Return(map[primitive.ObjectID]string{
			customerID: "สมชาย ใจดี",
		}, nil)

		result, err := service.Search(ctx, branchID, "สมชาย", 5)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "สมชาย ใจดี", result[0].CustomerName)
		mockPawnRepo.AssertExpectations(t)
		mockCustomerRepo.AssertExpectations(t)
	})

	t.Run("NoCustomerMatches", func(t *testing.T) {
		mockPawnRepo := new(testutils.MockPawnRepository)
		mockCustomerRepo := new(testutils.MockCustomerRepository)
		service := NewService(mockPawnRepo, nil, mockCustomerRepo)

		mockCustomerRepo.On("Search", ctx, "P-0042", 5).Return([]*entity.Customer{}, nil)
		mockPawnRepo.On("Search", ctx, branchID, "P-0042", []primitive.ObjectID{}, 5).Return([]*entity.Pawn{}, nil)

		result, err := service.Search(ctx, branchID, "P-0042", 5)

		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("CustomerSearchError", func(t *testing.T) {
		mockPawnRepo := new(testutils.MockPawnRepository)
		mockCustomerRepo := new(testutils.MockCustomerRepository)
		service := NewService(mockPawnRepo, nil, mockCustomerRepo)

		mockCustomerRepo.On("Search", ctx, "x", 5).Return([]*entity.Customer{}, errors.New("db down"))

		result, err := service.Search(ctx, branchID, "x", 5)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockPawnRepo.AssertNotCalled(t, "Search")
	})
}
