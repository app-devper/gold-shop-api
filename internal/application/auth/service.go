package auth

import (
	"context"
	"errors"

	"github.com/devper-gold/gold-shop-api/internal/domain/entity"
	"github.com/devper-gold/gold-shop-api/internal/domain/repository"
	"github.com/devper-gold/gold-shop-api/pkg/jwt"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service handles authentication logic
type Service struct {
	userRepo   repository.UserRepository
	jwtManager *jwt.Manager
}

// NewService creates a new Auth service
func NewService(userRepo repository.UserRepository, jwtManager *jwt.Manager) *Service {
	return &Service{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Login authenticates a user and returns a JWT token
func (s *Service) Login(ctx context.Context, username, password string) (string, *entity.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	if !user.IsActive {
		return "", nil, errors.New("user account is inactive")
	}

	if !utils.CheckPassword(password, user.PasswordHash) {
		return "", nil, errors.New("invalid credentials")
	}

	token, err := s.jwtManager.GenerateToken(user.ID, user.Username, string(user.Role), user.BranchID)
	if err != nil {
		return "", nil, errors.New("failed to generate token")
	}

	return token, user, nil
}

// RefreshToken generates a new token from existing claims
func (s *Service) RefreshToken(claims interface{}) (string, error) {
	jwtClaims, ok := claims.(*jwt.Claims)
	if !ok {
		return "", errors.New("invalid claims")
	}

	return s.jwtManager.RefreshToken(jwtClaims)
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(ctx context.Context, userID string) (*entity.User, error) {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, entity.ErrNotFound
	}

	return s.userRepo.GetByID(ctx, id)
}

// ChangePassword changes user password
func (s *Service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return entity.ErrNotFound
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !utils.CheckPassword(currentPassword, user.PasswordHash) {
		return errors.New("current password is incorrect")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user.PasswordHash = hashedPassword
	return s.userRepo.Update(ctx, user)
}
