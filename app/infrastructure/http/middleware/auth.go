package middleware

import (
	"github.com/app-devper/um-api/sessionclient"
	"github.com/app-devper/um-api/sessionclient/ginauth"
	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	mongoinfra "github.com/devper-gold/gold-shop-api/app/infrastructure/mongo"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// NewAuth verifies UM access tokens for gold-shop: SYSTEM binds the token and
// the session is confirmed in UM's Redis at redisHost (um-api ADR-0005).
// gold-shop serves many clients, so the token's client is not pinned. It
// fails when the secret, system, or Redis host is missing.
func NewAuth(secretKey, system, redisHost string) (*ginauth.Auth, error) {
	store, err := sessionclient.RedisStoreFor(redisHost)
	if err != nil {
		return nil, err
	}
	return NewAuthWithStore(secretKey, system, store)
}

// NewAuthWithStore is NewAuth with UM's session store supplied, for tests.
func NewAuthWithStore(secretKey, system string, store sessionclient.Store) (*ginauth.Auth, error) {
	verifier, err := sessionclient.NewVerifier(sessionclient.Config{SecretKey: secretKey, System: system, Store: store})
	if err != nil {
		return nil, err
	}
	return ginauth.New(verifier, func(c *gin.Context, e *sessionclient.Error) {
		utils.ErrorResponse(c, e.Code, e.Status, e.Message)
		c.Abort()
	}), nil
}

// RequireSession admits a caller with a live UM session. Every gold-shop
// route uses the default outage policy: while UM is unreachable a read may
// continue with the session last confirmed for its token, and writes wait.
func RequireSession(auth *ginauth.Auth) gin.HandlerFunc {
	return auth.Require(sessionclient.ReadOnlyWithLastGood)
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
