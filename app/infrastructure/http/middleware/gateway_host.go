package middleware

import (
	"github.com/app-devper/um-api/servicekit/gateway"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
)

// GatewayHostMiddleware refuses requests that did not come through the
// gateway (um-api servicekit, its ADR-0007), in this service's envelope.
func GatewayHostMiddleware(allowedHosts string) gin.HandlerFunc {
	hosts := gateway.ParseHosts(allowedHosts)
	return func(c *gin.Context) {
		if !hosts.Allows(c.Request) {
			utils.ForbiddenResponse(c, "GTW-403-001", gateway.Message)
			c.Abort()
			return
		}
		c.Next()
	}
}
