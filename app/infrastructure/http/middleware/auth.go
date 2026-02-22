package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SessionLookup is an interface for looking up sessions
type SessionLookup interface {
	GetSessionById(ctx context.Context, sessionId string) (string, error)
}

// AccessClaims represents JWT claims from um-api
type AccessClaims struct {
	Role     string `json:"role"`
	System   string `json:"system"`
	ClientId string `json:"clientId"`
	jwt.StandardClaims
}

// RequireAuthenticated validates the JWT token issued by um-api
func RequireAuthenticated(secretKey string) gin.HandlerFunc {
	jwtKey := []byte(secretKey)
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if !strings.HasPrefix(token, "Bearer ") {
			utils.UnauthorizedResponse(c, "missing authorization header")
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
			utils.UnauthorizedResponse(c, "token invalid")
			c.Abort()
			return
		}
		if tkn == nil || !tkn.Valid || claims.Id == "" {
			utils.UnauthorizedResponse(c, "token invalid")
			c.Abort()
			return
		}

		c.Set("SessionId", claims.Id)
		c.Set("Role", claims.Role)
		c.Set("System", claims.System)
		c.Set("ClientId", claims.ClientId)

		logrus.Info("SessionId: " + claims.Id)
		logrus.Info("Role: " + claims.Role)
		c.Next()
	}
}

// RequireSession validates the session in Redis and sets UserId in context
func RequireSession(sessionRepo SessionLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionId := c.GetString("SessionId")
		userId, err := sessionRepo.GetSessionById(c.Request.Context(), sessionId)
		if err != nil {
			utils.UnauthorizedResponse(c, "session invalid")
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
				utils.InternalErrorResponse(c, "failed to resolve employee")
				c.Abort()
				return
			}
			defaultBranch, bErr := branchRepo.GetByCode(ctx, "HQ")
			if bErr != nil {
				utils.ForbiddenResponse(c, "no branch available")
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
			utils.ForbiddenResponse(c, "Invalid request, restricted endpoint")
			c.Abort()
			return
		}
		for _, r := range allowedRoles {
			if string(r) == role {
				c.Next()
				return
			}
		}
		utils.ForbiddenResponse(c, "Don't have permission")
		c.Abort()
	}
}

func RequireRole(allowedRoles ...entity.UMRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("Role")
		if role == "" {
			utils.ForbiddenResponse(c, "Invalid request, restricted endpoint")
			c.Abort()
			return
		}
		for _, r := range allowedRoles {
			if string(r) == role {
				c.Next()
				return
			}
		}
		utils.ForbiddenResponse(c, "Don't have um permission")
		c.Abort()
	}
}
