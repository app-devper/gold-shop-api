package customer

import (
	"context"
	"errors"
	"time"

	"github.com/devper-gold/gold-shop-api/internal/domain/entity"
	"github.com/devper-gold/gold-shop-api/internal/domain/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	repo repository.CustomerRepository
}

func NewService(repo repository.CustomerRepository) *Service {
	return &Service{repo: repo}
}

type CreateCustomerRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	IDCard   string `json:"id_card"`
	Address  string `json:"address"`
	IsMember bool   `json:"is_member"`
	RFIDCard string `json:"rfid_card"`
	Email    string `json:"email"`
}

type UpdateCustomerRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	IDCard   string `json:"id_card"`
	Address  string `json:"address"`
	Email    string `json:"email"`
	RFIDCard string `json:"rfid_card"`
}

func (s *Service) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*entity.Customer, error) {
	// Check if exists by phone
	existing, _ := s.repo.GetByPhone(ctx, req.Phone)
	if existing != nil {
		return nil, errors.New("customer with this phone number already exists")
	}

	customer := entity.NewCustomer(req.FullName, req.Phone, req.IsMember)
	customer.IDCard = req.IDCard
	customer.Address = req.Address
	customer.RFIDCard = req.RFIDCard
	customer.Email = req.Email

	// Generate Member Code if member
	if req.IsMember {
		customer.MemberCode = generateMemberCode() // Simple generation
	}

	if err := s.repo.Create(ctx, customer); err != nil {
		return nil, err
	}

	return customer, nil
}

func (s *Service) GetCustomer(ctx context.Context, id string) (*entity.Customer, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid customer ID")
	}
	return s.repo.GetByID(ctx, objectID)
}

func (s *Service) GetCustomers(ctx context.Context, limit, offset int) ([]*entity.Customer, error) {
	return s.repo.GetAll(ctx, limit, offset)
}

func (s *Service) SearchCustomers(ctx context.Context, query string, limit int) ([]*entity.Customer, error) {
	return s.repo.Search(ctx, query, limit)
}

func (s *Service) UpdateCustomer(ctx context.Context, id string, req UpdateCustomerRequest) (*entity.Customer, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid customer ID")
	}

	customer, err := s.repo.GetByID(ctx, objectID)
	if err != nil {
		return nil, err
	}

	if req.FullName != "" {
		customer.FullName = req.FullName
	}
	if req.Phone != "" {
		customer.Phone = req.Phone
	}
	if req.IDCard != "" {
		customer.IDCard = req.IDCard
	}
	if req.Address != "" {
		customer.Address = req.Address
	}
	if req.Email != "" {
		customer.Email = req.Email
	}
	if req.RFIDCard != "" {
		customer.RFIDCard = req.RFIDCard
	}

	if err := s.repo.Update(ctx, customer); err != nil {
		return nil, err
	}

	return customer, nil
}

func (s *Service) DeleteCustomer(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid customer ID")
	}
	return s.repo.Delete(ctx, objectID)
}

// GetByRFID retrieves customer by RFID
func (s *Service) GetByRFID(ctx context.Context, rfid string) (*entity.Customer, error) {
	return s.repo.GetByRFID(ctx, rfid)
}

func generateMemberCode() string {
	// In real app, check DB for uniqueness or use sequence
	return "MB" + time.Now().Format("20060102150405")
}
