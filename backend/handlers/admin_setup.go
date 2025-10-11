package handlers

import (
	"civicweave/backend/models"
	"civicweave/backend/services"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// AdminSetupHandler handles admin setup operations
type AdminSetupHandler struct {
	userService  *models.UserService
	adminService *models.AdminService
	emailService *services.EmailService
}

// NewAdminSetupHandler creates a new AdminSetupHandler
func NewAdminSetupHandler(userService *models.UserService, adminService *models.AdminService, emailService *services.EmailService) *AdminSetupHandler {
	return &AdminSetupHandler{
		userService:  userService,
		adminService: adminService,
		emailService: emailService,
	}
}

// CreateAdminRequest represents the request to create an admin user
type CreateAdminRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password"` // Optional - will use ADMIN_PASSWORD env var if not provided
	Name     string `json:"name"`     // Optional - will use ADMIN_NAME env var if not provided
}

// CreateAdmin creates an admin user directly (bypasses email verification)
func (h *AdminSetupHandler) CreateAdmin(c *gin.Context) {
	log.Printf("👑 ADMIN_SETUP: Starting admin creation")

	var req CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ ADMIN_SETUP: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("📧 ADMIN_SETUP: Email: %s", req.Email)
	log.Printf("🔑 ADMIN_SETUP: Password provided: %v", req.Password != "")
	log.Printf("👤 ADMIN_SETUP: Name provided: %v", req.Name != "")

	// Check if user already exists
	log.Printf("🔍 ADMIN_SETUP: Checking if user already exists...")
	existingUser, err := h.userService.GetByEmail(req.Email)
	if err != nil {
		log.Printf("❌ ADMIN_SETUP: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if existingUser != nil {
		log.Printf("❌ ADMIN_SETUP: User already exists: %s", req.Email)
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}
	log.Printf("✅ ADMIN_SETUP: User does not exist, proceeding with creation")

	// Use password from environment variable if not provided in request
	password := req.Password
	if password == "" {
		password = os.Getenv("ADMIN_PASSWORD")
		if password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Admin password not configured. Set ADMIN_PASSWORD environment variable or provide password in request."})
			return
		}
	}

	// Trim whitespace (including newlines) from password
	password = strings.TrimSpace(password)

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Use name from environment variable if not provided in request
	name := req.Name
	if name == "" {
		name = os.Getenv("ADMIN_NAME")
		if name == "" {
			name = "System Administrator"
		}
	}

	// Create user with email verified
	user := &models.User{
		Email:         req.Email,
		PasswordHash:  string(hashedPassword),
		EmailVerified: true, // Skip email verification
		Role:          "admin",
	}

	if err := h.userService.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Create admin profile
	admin := &models.Admin{
		UserID: user.ID,
		Name:   name,
	}

	if err := h.adminService.Create(admin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin profile"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Admin user created successfully",
		"user_id": user.ID,
		"email":   user.Email,
		"name":    name,
	})
}
