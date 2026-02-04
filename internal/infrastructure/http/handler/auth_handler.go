package handler

import (
	"github.com/devper-gold/gold-shop-api/internal/application/auth"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *auth.Service
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// LoginRequest represents login request body
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	token, user, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		utils.UnauthorizedResponse(c, "Invalid credentials")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"token": token,
		"user":  user,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT system, logout is typically handled client-side
	// For stateful sessions, you'd invalidate the token here
	utils.MessageResponse(c, "Logged out successfully")
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		utils.UnauthorizedResponse(c, "No claims found")
		return
	}

	token, err := h.authService.RefreshToken(claims)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to refresh token")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"token": token,
	})
}

// GetCurrentUser returns the current authenticated user
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not found")
		return
	}

	user, err := h.authService.GetUserByID(c.Request.Context(), userID.(string))
	if err != nil {
		utils.NotFoundResponse(c, "User not found")
		return
	}

	utils.SuccessResponse(c, user)
}

// ChangePasswordRequest represents change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	userID, _ := c.Get("user_id")
	err := h.authService.ChangePassword(c.Request.Context(), userID.(string), req.CurrentPassword, req.NewPassword)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.MessageResponse(c, "Password changed successfully")
}
