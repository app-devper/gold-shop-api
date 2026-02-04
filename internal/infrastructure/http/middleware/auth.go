package middleware

import (
	"strings"

	"github.com/devper-gold/gold-shop-api/pkg/jwt"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.UnauthorizedResponse(c, "Authorization header required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.UnauthorizedResponse(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(parts[1])
		if err != nil {
			if err == jwt.ErrExpiredToken {
				utils.UnauthorizedResponse(c, "Token has expired")
			} else {
				utils.UnauthorizedResponse(c, "Invalid token")
			}
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("branch_id", claims.BranchID)
		c.Set("claims", claims)

		c.Next()
	}
}

// RoleMiddleware creates role-based authorization middleware
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			utils.UnauthorizedResponse(c, "Role not found in token")
			c.Abort()
			return
		}

		userRole := role.(string)
		allowed := false
		for _, r := range allowedRoles {
			if r == userRole {
				allowed = true
				break
			}
		}

		if !allowed {
			utils.ForbiddenResponse(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// PermissionMiddleware creates permission-based authorization middleware
func PermissionMiddleware(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			utils.UnauthorizedResponse(c, "Claims not found")
			c.Abort()
			return
		}

		// For now, check based on role
		// In production, you'd check against user's permissions array
		userClaims := claims.(*jwt.Claims)

		// Admin has all permissions
		if userClaims.Role == "admin" {
			c.Next()
			return
		}

		// TODO: Implement proper permission checking
		c.Next()
	}
}
