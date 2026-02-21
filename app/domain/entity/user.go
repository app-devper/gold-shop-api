package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRole represents user roles in the system
type UserRole string

const (
	RoleAdmin      UserRole = "admin"
	RoleManager    UserRole = "manager"
	RoleCashier    UserRole = "cashier"
	RoleAccountant UserRole = "accountant"
)

// User represents a system user
type User struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	BranchID     primitive.ObjectID `json:"branch_id" bson:"branch_id"`
	EmployeeCode string             `json:"employee_code" bson:"employee_code"`
	Username     string             `json:"username" bson:"username"`
	PasswordHash string             `json:"-" bson:"password_hash"`
	FullName     string             `json:"full_name" bson:"full_name"`
	Role         UserRole           `json:"role" bson:"role"`
	Permissions  []string           `json:"permissions" bson:"permissions"`
	IsActive     bool               `json:"is_active" bson:"is_active"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

// NewUser creates a new User entity
func NewUser(branchID primitive.ObjectID, employeeCode, username, passwordHash, fullName string, role UserRole) *User {
	now := time.Now()
	return &User{
		BranchID:     branchID,
		EmployeeCode: employeeCode,
		Username:     username,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Role:         role,
		Permissions:  GetDefaultPermissions(role),
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// GetDefaultPermissions returns default permissions based on role
func GetDefaultPermissions(role UserRole) []string {
	switch role {
	case RoleAdmin:
		return []string{
			"branch:*", "user:*", "product:*", "customer:*",
			"sale:*", "pawn:*", "gold_saving:*", "expense:*",
			"inventory:*", "report:*", "settings:*",
		}
	case RoleManager:
		return []string{
			"user:read", "product:*", "customer:*",
			"sale:*", "pawn:*", "gold_saving:*", "expense:*",
			"inventory:*", "report:*",
		}
	case RoleCashier:
		return []string{
			"product:read", "customer:*",
			"sale:*", "pawn:*", "gold_saving:*",
		}
	case RoleAccountant:
		return []string{
			"product:read", "customer:read",
			"sale:read", "pawn:read", "expense:*", "report:*",
		}
	default:
		return []string{}
	}
}

// HasPermission checks if user has a specific permission
func (u *User) HasPermission(permission string) bool {
	for _, p := range u.Permissions {
		if p == permission || p == "*" {
			return true
		}
		// Check wildcard permissions (e.g., "sale:*" matches "sale:create")
		if len(p) > 2 && p[len(p)-2:] == ":*" {
			prefix := p[:len(p)-1]
			if len(permission) > len(prefix) && permission[:len(prefix)] == prefix {
				return true
			}
		}
	}
	return false
}
