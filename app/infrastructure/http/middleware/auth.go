package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	mongoinfra "github.com/devper-gold/gold-shop-api/app/infrastructure/mongo"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// SessionLookup confirms the UM session behind a verified token and returns
// its user id (see redis.SessionRepository).
type SessionLookup interface {
	Authorize(ctx context.Context, sessionId, system, method string) (string, error)
}

// AccessClaims represents JWT claims from um-api
type AccessClaims struct {
	Role     string `json:"role"`
	System   string `json:"system"`
	ClientId string `json:"clientId"`
	jwt.RegisteredClaims
}

// RequireAuthenticated validates the JWT token issued by um-api and accepts
// only tokens issued for this service's system, so a live session from
// another system (for example POS) cannot call gold-shop.
func RequireAuthenticated(secretKey, system string) gin.HandlerFunc {
	jwtKey := []byte(secretKey)
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if !strings.HasPrefix(token, "Bearer ") {
			utils.UnauthorizedResponse(c, "AUT-401-001", "missing authorization header")
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(token, "Bearer ")
		claims := &AccessClaims{}
		tkn, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtKey, nil
		})
		if err != nil {
			utils.UnauthorizedResponse(c, "AUT-401-002", "token invalid")
			c.Abort()
			return
		}
		if tkn == nil || !tkn.Valid || claims.ID == "" {
			utils.UnauthorizedResponse(c, "AUT-401-003", "token invalid")
			c.Abort()
			return
		}
		if claims.System != system {
			utils.UnauthorizedResponse(c, "AUT-401-006", "system invalid")
			c.Abort()
			return
		}

		c.Set("SessionId", claims.ID)
		c.Set("Role", claims.Role)
		c.Set("System", claims.System)
		c.Set("ClientId", claims.ClientId)

		logrus.Info("SessionId: " + claims.ID)
		logrus.Info("Role: " + claims.Role)
		c.Next()
	}
}

func RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.GetString("ClientId")
		if err := mongoinfra.ValidateClientID(clientID); err != nil {
			logrus.Warnf("RequireTenant: rejecting request: %v", err)
			utils.UnauthorizedResponse(c, "AUT-401-004", "invalid tenant")
			c.Abort()
			return
		}
		ctx := mongoinfra.WithClientID(c.Request.Context(), clientID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// RequireSession validates the session in Redis and sets UserId in context
func RequireSession(sessionRepo SessionLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, err := sessionRepo.Authorize(c.Request.Context(),
			c.GetString("SessionId"), c.GetString("System"), c.Request.Method)
		if errors.Is(err, sessionclient.ErrUnavailable) {
			// Retry later; do not sign the user out.
			utils.ErrorResponse(c, "AUT-503-001", http.StatusServiceUnavailable, "identity service unavailable")
			c.Abort()
			return
		}
		if err != nil {
			utils.UnauthorizedResponse(c, "AUT-401-005", "session invalid")
			c.Abort()
			return
		}
		c.Set("UserId", userId)
		logrus.Info("UserId: " + userId)
		c.Next()
	}
}

// RequireBranch resolves the employee's branch from userId; falls back to HQ
func RequireBranch(employeeRepo repository.EmployeeRepository, branchRepo repository.BranchRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetString("UserId")
		ctx := c.Request.Context()

		employee, err := employeeRepo.GetByUserID(ctx, userId)
		if err != nil {
			if err != entity.ErrNotFound {
				utils.InternalErrorResponse(c, "AUT-500-001", "failed to resolve employee")
				c.Abort()
				return
			}
			defaultBranch, bErr := branchRepo.GetByCode(ctx, "HQ")
			if bErr != nil {
				utils.ForbiddenResponse(c, "AUT-403-001", "no branch available")
				c.Abort()
				return
			}
			c.Set("BranchId", defaultBranch.ID.Hex())
			c.Set("EmployeeRole", string(entity.EmployeeRoleStaff))
			logrus.Info("BranchId: " + defaultBranch.ID.Hex() + " (HQ fallback)")
		} else {
			c.Set("BranchId", employee.BranchID.Hex())
			c.Set("EmployeeRole", employee.Role)
			logrus.Info("BranchId: " + employee.BranchID.Hex())
			logrus.Info("EmployeeRole: " + employee.Role)
		}
		c.Next()
	}
}

func RoleMiddleware(allowedRoles ...entity.EmployeeRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("EmployeeRole")
		if role == "" {
			utils.ForbiddenResponse(c, "AUT-403-002", "Invalid request, restricted endpoint")
			c.Abort()
			return
		}
		for _, r := range allowedRoles {
			if string(r) == role {
				c.Next()
				return
			}
		}
		utils.ForbiddenResponse(c, "AUT-403-003", "Don't have permission")
		c.Abort()
	}
}

func RequireRole(allowedRoles ...entity.UMRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("Role")
		if role == "" {
			utils.ForbiddenResponse(c, "AUT-403-004", "Invalid request, restricted endpoint")
			c.Abort()
			return
		}
		for _, r := range allowedRoles {
			if string(r) == role {
				c.Next()
				return
			}
		}
		utils.ForbiddenResponse(c, "AUT-403-005", "Don't have um permission")
		c.Abort()
	}
}
