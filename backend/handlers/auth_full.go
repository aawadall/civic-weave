package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"civicweave/backend/config"
	"civicweave/backend/middleware"
	"civicweave/backend/models"
	"civicweave/backend/services"
	"civicweave/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	UserService         *models.UserService
	VolunteerService    *models.VolunteerService
	AdminService        *models.AdminService
	OAuthAccountService *models.OAuthAccountService
	EmailService        *services.EmailService
	GeocodingService    *utils.GeocodingService
	config              *config.Config
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	userService *models.UserService,
	volunteerService *models.VolunteerService,
	adminService *models.AdminService,
	oauthAccountService *models.OAuthAccountService,
	emailService *services.EmailService,
	geocodingService *utils.GeocodingService,
	config *config.Config,
) *AuthHandler {
	return &AuthHandler{
		UserService:         userService,
		VolunteerService:    volunteerService,
		AdminService:        adminService,
		OAuthAccountService: oauthAccountService,
		EmailService:        emailService,
		GeocodingService:    geocodingService,
		config:              config,
	}
}

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
	// Role is hardcoded to "volunteer" for regular registration
	// Admins can promote users later through admin interface
	Phone             string          `json:"phone"`
	LocationAddress   string          `json:"location_address"`
	SkillsDescription string          `json:"skills_description"`
	Availability      json.RawMessage `json:"availability"`
	SkillsVisible     bool            `json:"skills_visible"`
	ConsentGiven      bool            `json:"consent_given" binding:"required"`
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate consent
	if !req.ConsentGiven {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Consent must be given to register"})
		return
	}

	// Check if user already exists
	existingUser, err := h.UserService.GetByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	user := &models.User{
		Email:         req.Email,
		PasswordHash:  string(hashedPassword),
		EmailVerified: false,
		Role:          "volunteer", // Default role for regular registration
	}

	if err := h.UserService.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Geocode location if provided
	var locationLat, locationLng *float64
	if req.LocationAddress != "" {
		lat, lng, _, err := h.GeocodingService.GeocodeAddress(req.LocationAddress)
		if err != nil {
			// Log error but don't fail registration
			// Location can be added later
		} else {
			locationLat = &lat
			locationLng = &lng
		}
	}

	// Create volunteer profile (regular registration always creates volunteers)
	volunteer := &models.Volunteer{
		UserID:          user.ID,
		Name:            req.Name,
		Phone:           req.Phone,
		LocationLat:     locationLat,
		LocationLng:     locationLng,
		LocationAddress: req.LocationAddress,
		Skills:          []string{}, // Legacy field - will be replaced by skill claims
		Availability:    req.Availability,
		SkillsVisible:   req.SkillsVisible,
		ConsentGiven:    req.ConsentGiven,
	}

	if err := h.VolunteerService.Create(volunteer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create volunteer profile"})
		return
	}

	// Create skill claim if skills description is provided
	if req.SkillsDescription != "" && len(req.SkillsDescription) >= 10 {
		// TODO: This would require the embedding service to be available
		// For now, we'll skip creating the skill claim during registration
		// The user can add skill claims later through the profile page
	}

	// Generate verification token and send email
	verificationToken := utils.GenerateRandomToken()
	if err := h.createEmailVerificationToken(user.ID, verificationToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create verification token"})
		return
	}

	// Send verification email
	if err := h.EmailService.SendVerificationEmail(user.Email, verificationToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully. Please check your email for verification.",
		"user_id": user.ID,
	})
}

// LoginRequest represents user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	// Only log in development mode
	if gin.Mode() == gin.DebugMode {
		log.Printf("🔐 LOGIN: Starting login attempt")
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if gin.Mode() == gin.DebugMode {
			log.Printf("❌ LOGIN: Failed to bind JSON: %v", err)
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if gin.Mode() == gin.DebugMode {
		log.Printf("📧 LOGIN: Email: %s", req.Email)
		log.Printf("🔑 LOGIN: Password length: %d", len(req.Password))
		log.Printf("🔑 LOGIN: Password (masked): %s", maskPassword(req.Password))
	}

	// Get user by email
	if gin.Mode() == gin.DebugMode {
		log.Printf("🔍 LOGIN: Looking up user by email...")
	}
	user, err := h.UserService.GetByEmail(req.Email)
	if err != nil {
		if gin.Mode() == gin.DebugMode {
			log.Printf("❌ LOGIN: Database error: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if user == nil {
		if gin.Mode() == gin.DebugMode {
			log.Printf("❌ LOGIN: User not found for email: %s", req.Email)
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if gin.Mode() == gin.DebugMode {
		log.Printf("✅ LOGIN: User found - ID: %s, Role: %s, EmailVerified: %v", user.ID, user.Role, user.EmailVerified)
		log.Printf("🔑 LOGIN: Password hash present: %v", len(user.PasswordHash) > 0)
	}

	// Check if user has a password (not OAuth-only user)
	if user.PasswordHash == "" {
		if gin.Mode() == gin.DebugMode {
			log.Printf("❌ LOGIN: User has no password (OAuth-only user). Use Google Sign-In instead.")
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "This account uses Google Sign-In. Please use the Google Sign-In button."})
		return
	}

	// Check password
	if gin.Mode() == gin.DebugMode {
		log.Printf("🔍 LOGIN: Comparing passwords...")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		if gin.Mode() == gin.DebugMode {
			log.Printf("❌ LOGIN: Password comparison failed: %v", err)
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if gin.Mode() == gin.DebugMode {
		log.Printf("✅ LOGIN: Password comparison successful")
	}

	// Check if email is verified
	if !user.EmailVerified {
		if gin.Mode() == gin.DebugMode {
			log.Printf("❌ LOGIN: Email not verified for user: %s", user.Email)
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Please verify your email before logging in"})
		return
	}
	if gin.Mode() == gin.DebugMode {
		log.Printf("✅ LOGIN: Email verified")
	}

	// Generate JWT token
	if gin.Mode() == gin.DebugMode {
		log.Printf("🎫 LOGIN: Generating JWT token...")
	}
	token, err := middleware.GenerateJWT(user, h.config.JWT.Secret)
	if err != nil {
		if gin.Mode() == gin.DebugMode {
			log.Printf("❌ LOGIN: Failed to generate JWT token: %v", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	if gin.Mode() == gin.DebugMode {
		log.Printf("✅ LOGIN: JWT token generated successfully")
	}

	// Get user profile based on role
	if gin.Mode() == gin.DebugMode {
		log.Printf("👤 LOGIN: Getting user profile for role: %s", user.Role)
	}
	var userProfile interface{}
	if user.Role == "volunteer" {
		volunteer, err := h.VolunteerService.GetByUserID(user.ID)
		if err != nil {
			if gin.Mode() == gin.DebugMode {
				log.Printf("❌ LOGIN: Failed to get volunteer profile: %v", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get volunteer profile"})
			return
		}
		userProfile = volunteer
		if gin.Mode() == gin.DebugMode {
			log.Printf("✅ LOGIN: Volunteer profile retrieved")
		}
	} else if user.Role == "admin" {
		admin, err := h.AdminService.GetByUserID(user.ID)
		if err != nil {
			if gin.Mode() == gin.DebugMode {
				log.Printf("❌ LOGIN: Failed to get admin profile: %v", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get admin profile"})
			return
		}
		userProfile = admin
		if gin.Mode() == gin.DebugMode {
			log.Printf("✅ LOGIN: Admin profile retrieved")
		}
	}

	if gin.Mode() == gin.DebugMode {
		log.Printf("🎉 LOGIN: Login successful for user: %s", user.Email)
	}
	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  userProfile,
	})
}

// Helper function to mask password for logging
func maskPassword(password string) string {
	if len(password) == 0 {
		return ""
	}
	if len(password) <= 2 {
		return "**"
	}
	return password[:1] + strings.Repeat("*", len(password)-2) + password[len(password)-1:]
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// VerifyEmail handles email verification
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify token and mark email as verified
	userID, err := h.verifyEmailToken(req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired verification token"})
		return
	}

	// Mark email as verified
	if err := h.UserService.VerifyEmail(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
		return
	}

	// Get user and volunteer info for welcome email
	user, err := h.UserService.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	volunteer, err := h.VolunteerService.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get volunteer"})
		return
	}

	// Send welcome email
	if err := h.EmailService.SendWelcomeEmail(user.Email, volunteer.Name); err != nil {
		// Log error but don't fail verification
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get user profile based on role
	var userProfile interface{}
	var err error

	if userCtx.Role == "volunteer" {
		userProfile, err = h.VolunteerService.GetByUserID(userCtx.ID)
	} else if userCtx.Role == "admin" {
		userProfile, err = h.AdminService.GetByUserID(userCtx.ID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}

	c.JSON(http.StatusOK, userProfile)
}

// Helper methods for email verification
func (h *AuthHandler) createEmailVerificationToken(userID uuid.UUID, token string) error {
	// TODO: Implement database storage for verification tokens
	// For now, just return nil
	return nil
}

func (h *AuthHandler) verifyEmailToken(token string) (uuid.UUID, error) {
	// TODO: Implement token verification from database
	// For now, return a placeholder
	return uuid.New(), nil
}
