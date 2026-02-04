package user

import (
	"context"
	"errors"

	"github.com/devper-gold/gold-shop-api/internal/domain/entity"
	"github.com/devper-gold/gold-shop-api/internal/domain/repository"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateUserRequest represents data to create a user
type CreateUserRequest struct {
	BranchID     string `json:"branch_id" binding:"required"`
	EmployeeCode string `json:"employee_code" binding:"required"`
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
	FullName     string `json:"full_name" binding:"required"`
	Role         string `json:"role" binding:"required"`
}

// UpdateUserRequest represents data to update a user
type UpdateUserRequest struct {
	BranchID     string `json:"branch_id"`
	EmployeeCode string `json:"employee_code"`
	Username     string `json:"username"`
	FullName     string `json:"full_name"`
	Role         string `json:"role"`
	IsActive     *bool  `json:"is_active"`
}

// Service handles user management logic
type Service struct {
	userRepo   repository.UserRepository
	branchRepo repository.BranchRepository
}

// NewService creates a new User service
func NewService(userRepo repository.UserRepository, branchRepo repository.BranchRepository) *Service {
	return &Service{
		userRepo:   userRepo,
		branchRepo: branchRepo,
	}
}

// CreateUser creates a new user
func (s *Service) CreateUser(ctx context.Context, req *CreateUserRequest) (*entity.User, error) {
	// Check if username exists
	existingUser, _ := s.userRepo.GetByUsername(ctx, req.Username)
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// Validate Branch
	branchID, err := primitive.ObjectIDFromHex(req.BranchID)
	if err != nil {
		return nil, errors.New("invalid branch ID")
	}
	branch, err := s.branchRepo.GetByID(ctx, branchID)
	if err != nil || branch == nil {
		return nil, errors.New("branch not found")
	}

	// Hash Password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create User Entity
	role := entity.UserRole(req.Role)
	user := entity.NewUser(
		branchID,
		req.EmployeeCode,
		req.Username,
		hashedPassword,
		req.FullName,
		role,
	)

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *Service) GetUser(ctx context.Context, id string) (*entity.User, error) {
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetUsers retrieves all users
func (s *Service) GetUsers(ctx context.Context) ([]*entity.User, error) {
	return s.userRepo.GetAll(ctx)
}

// UpdateUser updates a user
func (s *Service) UpdateUser(ctx context.Context, id string, req *UpdateUserRequest) (*entity.User, error) {
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Update fields if provided
	if req.Username != "" && req.Username != user.Username {
		// Check uniqueness
		existing, _ := s.userRepo.GetByUsername(ctx, req.Username)
		if existing != nil && existing.ID != user.ID {
			return nil, errors.New("username already exists")
		}
		user.Username = req.Username
	}

	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.EmployeeCode != "" {
		user.EmployeeCode = req.EmployeeCode
	}
	if req.Role != "" {
		newRole := entity.UserRole(req.Role)
		user.Role = newRole
		user.Permissions = entity.GetDefaultPermissions(newRole)
	}
	if req.BranchID != "" {
		branchID, err := primitive.ObjectIDFromHex(req.BranchID)
		if err == nil {
			user.BranchID = branchID
		}
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes a user (soft delete usually, but repo has Delete)
func (s *Service) DeleteUser(ctx context.Context, id string) error {
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user ID")
	}

	return s.userRepo.Delete(ctx, userID)
}

// ResetPassword resets a user's password
func (s *Service) ResetPassword(ctx context.Context, id, newPassword string) error {
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user ID")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hashed
	return s.userRepo.Update(ctx, user)
}
