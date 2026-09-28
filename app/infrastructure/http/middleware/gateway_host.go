package middleware

import (
	"github.com/app-devper/um-api/servicekit/gateway"
	"github.com/app-devper/um-api/servicekit/gateway/gingateway"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
)

// GatewayHostMiddleware refuses requests that did not come through the
// gateway (um-api servicekit, its ADR-0007), in this service's envelope.
func GatewayHostMiddleware(allowedHosts string) gin.HandlerFunc {
	return gingateway.Middleware(gateway.ParseHosts(allowedHosts), func(c *gin.Context) {
		utils.ForbiddenResponse(c, "GTW-403-001", gateway.Message)
	})
}
