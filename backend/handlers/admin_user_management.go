package handlers

import (
	"log"
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

	// Check for constraints that would prevent deletion
	constraintIssues, err := h.checkUserDeletionConstraints(userID)
	if err != nil {
		log.Printf("❌ USER_DELETE_CONSTRAINT_CHECK: Failed to check constraints for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate user deletion constraints"})
		return
	}

	if len(constraintIssues) > 0 {
		log.Printf("❌ USER_DELETE_CONSTRAINTS: User %s has constraint violations: %v", userID, constraintIssues)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "Cannot delete user due to existing references",
			"constraints": constraintIssues,
		})
		return
	}

	// Delete user (this will cascade delete related records due to foreign key constraints)
	log.Printf("🔄 USER_DELETE: Attempting to delete user %s", userID)
	if err := h.userService.Delete(userID); err != nil {
		log.Printf("❌ USER_DELETE: Failed to delete user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user", "details": err.Error()})
		return
	}

	log.Printf("✅ USER_DELETE: Successfully deleted user %s", userID)

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

// checkUserDeletionConstraints checks for foreign key constraints that would prevent user deletion
func (h *AdminUserManagementHandler) checkUserDeletionConstraints(userID uuid.UUID) ([]string, error) {
	var constraintIssues []string

	// Check if user is team lead on any projects
	var teamLeadCount int
	err := h.userService.GetDB().QueryRow(`
		SELECT COUNT(*) FROM projects WHERE team_lead_id = $1
	`, userID).Scan(&teamLeadCount)
	if err != nil {
		return nil, err
	}
	if teamLeadCount > 0 {
		constraintIssues = append(constraintIssues, "User is team lead on active projects")
	}

	// Check if user has created any campaigns
	var campaignCount int
	err = h.userService.GetDB().QueryRow(`
		SELECT COUNT(*) FROM campaigns WHERE created_by_user_id = $1
	`, userID).Scan(&campaignCount)
	if err != nil {
		return nil, err
	}
	if campaignCount > 0 {
		constraintIssues = append(constraintIssues, "User has created campaigns")
	}

	// Check if user has created any project tasks
	var taskCount int
	err = h.userService.GetDB().QueryRow(`
		SELECT COUNT(*) FROM project_tasks WHERE created_by_id = $1
	`, userID).Scan(&taskCount)
	if err != nil {
		return nil, err
	}
	if taskCount > 0 {
		constraintIssues = append(constraintIssues, "User has created project tasks")
	}

	// Check if user has sent any project messages
	var messageCount int
	err = h.userService.GetDB().QueryRow(`
		SELECT COUNT(*) FROM project_messages WHERE sender_id = $1
	`, userID).Scan(&messageCount)
	if err != nil {
		return nil, err
	}
	if messageCount > 0 {
		constraintIssues = append(constraintIssues, "User has sent project messages")
	}

	return constraintIssues, nil
}
