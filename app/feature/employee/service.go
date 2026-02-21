package employee

import (
	"context"
	"errors"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateEmployeeRequest represents data to create an employee
type CreateEmployeeRequest struct {
	BranchID string `json:"branchId" binding:"required"`
	UserID   string `json:"userId" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

// UpdateEmployeeRequest represents data to update an employee
type UpdateEmployeeRequest struct {
	BranchID string `json:"branchId" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

// Service handles employee management logic
type Service struct {
	employeeRepo repository.EmployeeRepository
	branchRepo   repository.BranchRepository
}

// NewService creates a new Employee service
func NewService(employeeRepo repository.EmployeeRepository, branchRepo repository.BranchRepository) *Service {
	return &Service{
		employeeRepo: employeeRepo,
		branchRepo:   branchRepo,
	}
}

// CreateEmployee creates a new employee
func (s *Service) CreateEmployee(ctx context.Context, req *CreateEmployeeRequest, createdBy string) (*entity.Employee, error) {
	branchID, err := primitive.ObjectIDFromHex(req.BranchID)
	if err != nil {
		return nil, errors.New("invalid branch ID")
	}

	_, err = s.branchRepo.GetByID(ctx, branchID)
	if err != nil {
		return nil, errors.New("branch not found")
	}

	emp := &entity.Employee{
		BranchID:    branchID,
		UserID:      req.UserID,
		Role:        req.Role,
		CreatedBy:   createdBy,
		CreatedDate: time.Now(),
		UpdatedBy:   createdBy,
		UpdatedDate: time.Now(),
	}

	if err := s.employeeRepo.Create(ctx, emp); err != nil {
		if err == entity.ErrDuplicateKey {
			return nil, errors.New("employee with this userId already exists")
		}
		return nil, err
	}
	return emp, nil
}

// GetEmployee retrieves an employee by ID
func (s *Service) GetEmployee(ctx context.Context, id string) (*entity.Employee, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid employee ID")
	}
	return s.employeeRepo.GetByID(ctx, objID)
}

// GetEmployees retrieves all employees
func (s *Service) GetEmployees(ctx context.Context) ([]*entity.Employee, error) {
	return s.employeeRepo.GetAll(ctx)
}

// GetEmployeesByBranchID retrieves employees by branch ID
func (s *Service) GetEmployeesByBranchID(ctx context.Context, branchID string) ([]*entity.Employee, error) {
	objID, err := primitive.ObjectIDFromHex(branchID)
	if err != nil {
		return nil, errors.New("invalid branch ID")
	}
	return s.employeeRepo.GetByBranchID(ctx, objID)
}

// UpdateEmployee updates an employee
func (s *Service) UpdateEmployee(ctx context.Context, id string, req *UpdateEmployeeRequest, updatedBy string) (*entity.Employee, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid employee ID")
	}

	emp, err := s.employeeRepo.GetByID(ctx, objID)
	if err != nil {
		return nil, err
	}

	branchID, err := primitive.ObjectIDFromHex(req.BranchID)
	if err != nil {
		return nil, errors.New("invalid branch ID")
	}

	emp.BranchID = branchID
	emp.Role = req.Role
	emp.UpdatedBy = updatedBy
	emp.UpdatedDate = time.Now()

	if err := s.employeeRepo.Update(ctx, emp); err != nil {
		return nil, err
	}
	return emp, nil
}

// DeleteEmployee deletes an employee
func (s *Service) DeleteEmployee(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid employee ID")
	}
	return s.employeeRepo.Delete(ctx, objID)
}
