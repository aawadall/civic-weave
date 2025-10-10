package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AdminProfileHandler handles admin profile related requests
type AdminProfileHandler struct {
	db *sql.DB
}

// NewAdminProfileHandler creates a new AdminProfileHandler
func NewAdminProfileHandler(db *sql.DB) *AdminProfileHandler {
	return &AdminProfileHandler{db: db}
}

// AdminProfile represents the admin profile data
type AdminProfile struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	Name            string                 `json:"name" db:"name"`
	Email           string                 `json:"email" db:"email"`
	Phone           string                 `json:"phone" db:"phone"`
	LocationAddress string                 `json:"location_address" db:"location_address"`
	Preferences     map[string]interface{} `json:"preferences" db:"preferences"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// SystemStats represents system statistics for admin dashboard
type SystemStats struct {
	TotalVolunteers     int            `json:"total_volunteers"`
	ActiveInitiatives   int            `json:"active_initiatives"`
	PendingApplications int            `json:"pending_applications"`
	TotalSkillClaims    int            `json:"total_skill_claims"`
	LowWeightClaims     int            `json:"low_weight_claims"`
	RecentActivity      []ActivityItem `json:"recent_activity"`
}

// ActivityItem represents a recent system activity
type ActivityItem struct {
	Description string `json:"description"`
	Timestamp   string `json:"timestamp"`
	Type        string `json:"type"`
}

// GetAdminProfile handles GET /api/admin/profile
func (h *AdminProfileHandler) GetAdminProfile(c *gin.Context) {
	adminID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin ID not found in context"})
		return
	}

	var profile AdminProfile
	query := `
		SELECT id, name, email, phone, location_address, preferences, created_at, updated_at
		FROM admins 
		WHERE id = $1`

	var preferencesJSON []byte
	err := h.db.QueryRow(query, adminID).Scan(
		&profile.ID, &profile.Name, &profile.Email, &profile.Phone,
		&profile.LocationAddress, &preferencesJSON, &profile.CreatedAt, &profile.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin profile not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve admin profile"})
		return
	}

	// Parse preferences JSON
	if len(preferencesJSON) > 0 {
		if err := json.Unmarshal(preferencesJSON, &profile.Preferences); err != nil {
			// If parsing fails, use default preferences
			profile.Preferences = map[string]interface{}{
				"emailNotifications":   true,
				"skillReviewReminders": true,
				"weeklyReports":        false,
				"instantAlerts":        true,
			}
		}
	} else {
		// Default preferences if none stored
		profile.Preferences = map[string]interface{}{
			"emailNotifications":   true,
			"skillReviewReminders": true,
			"weeklyReports":        false,
			"instantAlerts":        true,
		}
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateAdminProfile handles PUT /api/admin/profile
func (h *AdminProfileHandler) UpdateAdminProfile(c *gin.Context) {
	adminID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin ID not found in context"})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	for key, value := range updateData {
		switch key {
		case "name", "phone", "location_address":
			setParts = append(setParts, key+" = $"+fmt.Sprintf("%d", argIndex))
			args = append(args, value)
			argIndex++
		case "preferences":
			preferencesJSON, err := json.Marshal(value)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid preferences format"})
				return
			}
			setParts = append(setParts, key+" = $"+fmt.Sprintf("%d", argIndex))
			args = append(args, preferencesJSON)
			argIndex++
		}
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	// Add updated_at
	setParts = append(setParts, "updated_at = $"+fmt.Sprintf("%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add admin ID as last argument
	args = append(args, adminID)

	query := "UPDATE admins SET " + strings.Join(setParts, ", ") + " WHERE id = $" + fmt.Sprintf("%d", argIndex)

	_, err := h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update admin profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// GetSystemStats handles GET /api/admin/stats
func (h *AdminProfileHandler) GetSystemStats(c *gin.Context) {
	stats := SystemStats{}

	// Get total volunteers
	err := h.db.QueryRow("SELECT COUNT(*) FROM volunteers").Scan(&stats.TotalVolunteers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get volunteer count"})
		return
	}

	// Get active initiatives
	err = h.db.QueryRow("SELECT COUNT(*) FROM initiatives WHERE status = 'active'").Scan(&stats.ActiveInitiatives)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active initiatives count"})
		return
	}

	// Get pending applications
	err = h.db.QueryRow("SELECT COUNT(*) FROM applications WHERE status = 'pending'").Scan(&stats.PendingApplications)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pending applications count"})
		return
	}

	// Get total skill claims
	err = h.db.QueryRow("SELECT COUNT(*) FROM skill_claims WHERE is_active = true").Scan(&stats.TotalSkillClaims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get skill claims count"})
		return
	}

	// Get low weight skill claims (weight < 0.4)
	err = h.db.QueryRow(`
		SELECT COUNT(*) 
		FROM skill_claims 
		WHERE is_active = true AND claim_weight < 0.4
	`).Scan(&stats.LowWeightClaims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get low weight claims count"})
		return
	}

	// Get recent activity (last 10 activities)
	activityQuery := `
		SELECT 
			CASE 
				WHEN 'volunteer_registration' THEN 'New volunteer registered: ' || v.name
				WHEN 'skill_claim' THEN 'New skill claim added: ' || sc.skill_name
				WHEN 'initiative_created' THEN 'New initiative created: ' || i.title
				ELSE 'System activity'
			END as description,
			COALESCE(v.created_at, sc.created_at, i.created_at) as timestamp
		FROM (
			SELECT 'volunteer_registration' as type, id, name, created_at, NULL as skill_name, NULL as title
			FROM volunteers 
			WHERE created_at >= NOW() - INTERVAL '7 days'
			UNION ALL
			SELECT 'skill_claim', sc.id, v.name, sc.created_at, sc.skill_name, NULL
			FROM skill_claims sc
			JOIN volunteers v ON sc.volunteer_id = v.id
			WHERE sc.created_at >= NOW() - INTERVAL '7 days'
			UNION ALL
			SELECT 'initiative_created', i.id, NULL, i.created_at, NULL, i.title
			FROM initiatives i
			WHERE i.created_at >= NOW() - INTERVAL '7 days'
		) as activities
		ORDER BY timestamp DESC
		LIMIT 10`

	rows, err := h.db.Query(activityQuery)
	if err != nil {
		// If activity query fails, just return empty activity list
		stats.RecentActivity = []ActivityItem{}
	} else {
		defer rows.Close()

		for rows.Next() {
			var activity ActivityItem
			err := rows.Scan(&activity.Description, &activity.Timestamp)
			if err != nil {
				continue
			}

			// Format timestamp
			if parsedTime, err := time.Parse(time.RFC3339, activity.Timestamp); err == nil {
				activity.Timestamp = parsedTime.Format("Jan 2, 3:04 PM")
			}

			stats.RecentActivity = append(stats.RecentActivity, activity)
		}
	}

	c.JSON(http.StatusOK, stats)
}

// ChangePassword handles PUT /api/admin/change-password
func (h *AdminProfileHandler) ChangePassword(c *gin.Context) {
	adminID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin ID not found in context"})
		return
	}

	var request struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify current password
	var hashedPassword string
	err := h.db.QueryRow("SELECT password_hash FROM admins WHERE id = $1", adminID).Scan(&hashedPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify current password"})
		return
	}

	// Verify current password
	if !verifyPassword(request.CurrentPassword, hashedPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password
	newHashedPassword, err := hashPassword(request.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	// Update password
	_, err = h.db.Exec("UPDATE admins SET password_hash = $1, updated_at = $2 WHERE id = $3",
		newHashedPassword, time.Now(), adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// Helper functions for password hashing
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
