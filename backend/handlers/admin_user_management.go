package handlers

import (
	"net/http"
	"time"

	"civicweave/backend/middleware"
	"civicweave/backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AdminUserManagementHandler handles admin user management requests
type AdminUserManagementHandler struct {
	userService      *models.UserService
	volunteerService *models.VolunteerService
	adminService     *models.AdminService
	roleService      *models.RoleService
}

// NewAdminUserManagementHandler creates a new AdminUserManagementHandler
func NewAdminUserManagementHandler(
	userService *models.UserService,
	volunteerService *models.VolunteerService,
	adminService *models.AdminService,
	roleService *models.RoleService,
) *AdminUserManagementHandler {
	return &AdminUserManagementHandler{
		userService:      userService,
		volunteerService: volunteerService,
		adminService:     adminService,
		roleService:      roleService,
	}
}

// DeleteUser handles DELETE /api/admin/users/:id
func (h *AdminUserManagementHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get admin ID from context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Prevent admin from deleting themselves
	if userCtx.ID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	// Check if user exists
	user, err := h.userService.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if user is an admin - prevent deletion of other admins
	hasAdminRole, err := h.userService.HasRole(userID, "admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user roles"})
		return
	}
	if hasAdminRole {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete admin users"})
		return
	}

	// Delete user (this will cascade delete related records due to foreign key constraints)
	if err := h.userService.Delete(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ForceVerificationStatus handles PUT /api/admin/users/:id/verification
func (h *AdminUserManagementHandler) ForceVerificationStatus(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		EmailVerified bool `json:"email_verified" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user exists
	user, err := h.userService.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update user verification status
	user.EmailVerified = req.EmailVerified
	if err := h.userService.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user verification status"})
		return
	}

	status := "verified"
	if !req.EmailVerified {
		status = "unverified"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "User verification status updated successfully",
		"email_verified": req.EmailVerified,
		"status":         status,
	})
}

// ChangeUserPassword handles PUT /api/admin/users/:id/password
func (h *AdminUserManagementHandler) ChangeUserPassword(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user exists
	user, err := h.userService.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update user password
	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()
	if err := h.userService.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User password changed successfully"})
}

// GetUserDetails handles GET /api/admin/users/:id
func (h *AdminUserManagementHandler) GetUserDetails(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get user with roles
	userWithRoles, err := h.userService.GetUserWithRoles(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user details"})
		return
	}
	if userWithRoles == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get user name from volunteer or admin profile
	var name string
	volunteer, err := h.volunteerService.GetByUserID(userID)
	if err == nil && volunteer != nil {
		name = volunteer.Name
	} else {
		admin, err := h.adminService.GetByUserID(userID)
		if err == nil && admin != nil {
			name = admin.Name
		}
	}

	// Create response with user details
	response := gin.H{
		"id":             userWithRoles.ID,
		"email":          userWithRoles.Email,
		"email_verified": userWithRoles.EmailVerified,
		"name":           name,
		"roles":          userWithRoles.Roles,
		"created_at":     userWithRoles.CreatedAt,
		"updated_at":     userWithRoles.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}
