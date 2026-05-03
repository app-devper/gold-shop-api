package reward

import (
	"context"
	"testing"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/testutils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// stubRewardRepo / stubRedemptionRepo are minimal in-memory fakes used here because
// testutils does not yet ship reward mocks.

type stubRewardRepo struct{ reward *entity.Reward }

func (s *stubRewardRepo) Create(ctx context.Context, r *entity.Reward) error { return nil }
func (s *stubRewardRepo) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Reward, error) {
	return s.reward, nil
}
func (s *stubRewardRepo) GetAll(ctx context.Context, activeOnly bool) ([]*entity.Reward, error) {
	return nil, nil
}
func (s *stubRewardRepo) Update(ctx context.Context, r *entity.Reward) error { return nil }
func (s *stubRewardRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	return nil
}

type stubRedemptionRepo struct{}

func (s *stubRedemptionRepo) Create(ctx context.Context, r *entity.RewardRedemption) error {
	return nil
}
func (s *stubRedemptionRepo) GetByCustomerID(ctx context.Context, id primitive.ObjectID) ([]*entity.RewardRedemption, error) {
	return nil, nil
}
func (s *stubRedemptionRepo) GetByBranchID(ctx context.Context, id primitive.ObjectID, limit, offset int) ([]*entity.RewardRedemption, error) {
	return nil, nil
}

func TestRedeemReward_NonMember_ReturnsErrorWithoutPanic(t *testing.T) {
	ctx := context.Background()
	customerID := primitive.NewObjectID()

	mockCustomer := new(testutils.MockCustomerRepository)
	mockCustomer.On("GetByID", ctx, customerID).Return(&entity.Customer{
		ID:         customerID,
		FullName:   "Walk-in",
		IsMember:   false,
		Membership: nil,
	}, nil)

	svc := NewService(&stubRewardRepo{}, &stubRedemptionRepo{}, mockCustomer, &testutils.MockTransactionManager{})

	_, err := svc.RedeemReward(ctx, customerID, primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID())
	if err == nil {
		t.Fatal("expected error for non-member, got nil")
	}
	if err.Error() != "customer is not a member" {
		t.Fatalf("expected 'customer is not a member', got %q", err.Error())
	}
	mockCustomer.AssertExpectations(t)
}
