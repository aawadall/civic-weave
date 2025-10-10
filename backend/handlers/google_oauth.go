package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"civicweave/backend/config"
	"civicweave/backend/middleware"
	"civicweave/backend/models"
	"civicweave/backend/services"

	"github.com/gin-gonic/gin"
)

// GoogleOAuthHandler handles Google OAuth authentication
type GoogleOAuthHandler struct {
	UserService         *models.UserService
	VolunteerService    *models.VolunteerService
	AdminService        *models.AdminService
	OAuthAccountService *models.OAuthAccountService
	EmailService        *services.EmailService
	config              *config.Config
}

// NewGoogleOAuthHandler creates a new Google OAuth handler
func NewGoogleOAuthHandler(
	userService *models.UserService,
	volunteerService *models.VolunteerService,
	adminService *models.AdminService,
	oauthAccountService *models.OAuthAccountService,
	emailService *services.EmailService,
	config *config.Config,
) *GoogleOAuthHandler {
	return &GoogleOAuthHandler{
		UserService:         userService,
		VolunteerService:    volunteerService,
		AdminService:        adminService,
		OAuthAccountService: oauthAccountService,
		EmailService:        emailService,
		config:              config,
	}
}

// GoogleAuthRequest represents Google OAuth request
type GoogleAuthRequest struct {
	Credential string `json:"credential" binding:"required"`
}

// GoogleUserInfo represents Google user information
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	EmailVerified bool   `json:"email_verified"`
}

// GoogleAuth handles Google OAuth authentication
func (h *GoogleOAuthHandler) GoogleAuth(c *gin.Context) {
	var req GoogleAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Verify the Google credential and get user info
	googleUser, err := h.verifyGoogleCredential(req.Credential)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Google credential"})
		return
	}

	// Check if user already exists by email
	user, err := h.UserService.GetByEmail(googleUser.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// If user doesn't exist, create a new one
	if user == nil {
		user, err = h.createUserFromGoogle(googleUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	}

	// Check if OAuth account is linked
	oauthAccount, err := h.OAuthAccountService.GetByProviderAndUserID("google", googleUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// If OAuth account doesn't exist, create it
	if oauthAccount == nil {
		oauthAccount = &models.OAuthAccount{
			UserID:         user.ID,
			Provider:       "google",
			ProviderUserID: googleUser.ID,
		}
		if err := h.OAuthAccountService.Create(oauthAccount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link OAuth account"})
			return
		}
	}

	// Generate JWT token
	token, err := middleware.GenerateJWT(user, h.config.JWT.Secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Get user profile based on role
	var userProfile interface{}
	if user.Role == "volunteer" {
		volunteer, err := h.VolunteerService.GetByUserID(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get volunteer profile"})
			return
		}
		userProfile = volunteer
	} else if user.Role == "admin" {
		admin, err := h.AdminService.GetByUserID(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get admin profile"})
			return
		}
		userProfile = admin
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  userProfile,
	})
}

// verifyGoogleCredential verifies the Google credential and returns user info
func (h *GoogleOAuthHandler) verifyGoogleCredential(credential string) (*GoogleUserInfo, error) {
	// For now, we'll decode the JWT credential directly
	// In production, you should verify the signature with Google's public keys
	
	// Split the JWT token
	parts := strings.Split(credential, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Decode the payload (second part)
	payload := parts[1]
	// Add padding if needed
	for len(payload)%4 != 0 {
		payload += "="
	}

	// Decode base64
	payloadBytes, err := base64URLDecode(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %v", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %v", err)
	}

	// Extract user info from claims
	googleUser := &GoogleUserInfo{
		ID:    getString(claims, "sub"),
		Email: getString(claims, "email"),
		Name:  getString(claims, "name"),
		Picture: getString(claims, "picture"),
		EmailVerified: getBool(claims, "email_verified"),
	}

	if googleUser.ID == "" || googleUser.Email == "" {
		return nil, fmt.Errorf("missing required user information")
	}

	return googleUser, nil
}

// createUserFromGoogle creates a new user from Google OAuth info
func (h *GoogleOAuthHandler) createUserFromGoogle(googleUser *GoogleUserInfo) (*models.User, error) {
	// For Google OAuth users, we'll default to volunteer role
	// Admins can be promoted later through the admin setup endpoint
	user := &models.User{
		Email:         googleUser.Email,
		PasswordHash:  "", // No password for OAuth users
		EmailVerified: googleUser.EmailVerified,
		Role:          "volunteer",
	}

	if err := h.UserService.Create(user); err != nil {
		return nil, err
	}

	// Create volunteer profile
	volunteer := &models.Volunteer{
		UserID:          user.ID,
		Name:            googleUser.Name,
		Phone:           "",
		LocationAddress: "",
		Skills:          []string{}, // Will be populated via skill claims
		Availability:    json.RawMessage(`{}`),
		SkillsVisible:   true,
		ConsentGiven:    true, // Implicit consent via Google OAuth
	}

	if err := h.VolunteerService.Create(volunteer); err != nil {
		return nil, err
	}

	return user, nil
}

// Helper functions
func base64URLDecode(str string) ([]byte, error) {
	// Replace URL-safe base64 characters
	str = strings.ReplaceAll(str, "-", "+")
	str = strings.ReplaceAll(str, "_", "/")
	
	// Add padding if needed
	for len(str)%4 != 0 {
		str += "="
	}
	
	return base64.StdEncoding.DecodeString(str)
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}
