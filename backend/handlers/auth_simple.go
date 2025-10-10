package handlers

import (
	"net/http"

	"civicweave/backend/config"
	"civicweave/backend/middleware"
	"civicweave/backend/models"

	"github.com/gin-gonic/gin"
)

// SimpleAuthHandler handles authentication requests (simplified for MVP)
type SimpleAuthHandler struct {
	UserService *models.UserService
	config      *config.Config
}

// NewSimpleAuthHandler creates a new simple auth handler
func NewSimpleAuthHandler(userService *models.UserService, config *config.Config) *SimpleAuthHandler {
	return &SimpleAuthHandler{
		UserService: userService,
		config:      config,
	}
}

// Register handles user registration (placeholder)
func (h *SimpleAuthHandler) Register(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Registration not implemented yet"})
}

// Login handles user login (placeholder)
func (h *SimpleAuthHandler) Login(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Login not implemented yet"})
}

// GoogleAuth handles Google OAuth (placeholder)
func (h *SimpleAuthHandler) GoogleAuth(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Google OAuth not implemented yet"})
}

// VerifyEmail handles email verification (placeholder)
func (h *SimpleAuthHandler) VerifyEmail(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Email verification not implemented yet"})
}

// ForgotPassword handles password reset request (placeholder)
func (h *SimpleAuthHandler) ForgotPassword(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Password reset not implemented yet"})
}

// ResetPassword handles password reset (placeholder)
func (h *SimpleAuthHandler) ResetPassword(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Password reset not implemented yet"})
}

// GetProfile returns the current user's profile (placeholder)
func (h *SimpleAuthHandler) GetProfile(c *gin.Context) {
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userCtx.ID,
		"email":   userCtx.Email,
		"role":    userCtx.Role,
	})
}

// UpdateProfile updates the current user's profile (placeholder)
func (h *SimpleAuthHandler) UpdateProfile(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Profile update not implemented yet"})
}
