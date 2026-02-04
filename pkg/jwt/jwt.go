package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// Claims represents JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	BranchID string `json:"branch_id"`
	jwt.RegisteredClaims
}

// Manager handles JWT operations
type Manager struct {
	secret          []byte
	expirationHours int
}

// NewManager creates a new JWT manager
func NewManager(secret string, expirationHours int) *Manager {
	return &Manager{
		secret:          []byte(secret),
		expirationHours: expirationHours,
	}
}

// GenerateToken generates a new JWT token
func (m *Manager) GenerateToken(userID primitive.ObjectID, username, role string, branchID primitive.ObjectID) (string, error) {
	claims := &Claims{
		UserID:   userID.Hex(),
		Username: username,
		Role:     role,
		BranchID: branchID.Hex(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(m.expirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ValidateToken validates a JWT token and returns claims
func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken generates a new token with extended expiration
func (m *Manager) RefreshToken(claims *Claims) (string, error) {
	userID, _ := primitive.ObjectIDFromHex(claims.UserID)
	branchID, _ := primitive.ObjectIDFromHex(claims.BranchID)
	return m.GenerateToken(userID, claims.Username, claims.Role, branchID)
}
