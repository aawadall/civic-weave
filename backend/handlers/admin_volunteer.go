package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"civicweave/backend/models"
)

// AdminVolunteerHandler handles admin operations on volunteers
type AdminVolunteerHandler struct {
	db                *sql.DB
	taxonomyService   *models.SkillTaxonomyService
}

// NewAdminVolunteerHandler creates a new admin volunteer handler
func NewAdminVolunteerHandler(db *sql.DB, taxonomyService *models.SkillTaxonomyService) *AdminVolunteerHandler {
	return &AdminVolunteerHandler{
		db:              db,
		taxonomyService: taxonomyService,
	}
}

// AdjustVolunteerSkillWeight handles PUT /api/admin/volunteers/:volunteer_id/skills/:skill_id/weight
func (h *AdminVolunteerHandler) AdjustVolunteerSkillWeight(c *gin.Context) {
	type Request struct {
		NewWeight     float64     `json:"new_weight" binding:"required,min=0.1,max=1.0"`
		Reason        string      `json:"reason"`
		InitiativeID  *uuid.UUID  `json:"initiative_id"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	volunteerIDStr := c.Param("volunteer_id")
	volunteerID, err := uuid.Parse(volunteerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid volunteer ID"})
		return
	}

	skillIDStr := c.Param("skill_id")
	skillID, err := strconv.Atoi(skillIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid skill ID"})
		return
	}

	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin ID not found"})
		return
	}

	adminUUID, ok := adminID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid admin ID"})
		return
	}

	// Get current weight
	var currentWeight float64
	err = h.db.QueryRow(`
		SELECT skill_weight FROM volunteer_skills 
		WHERE volunteer_id = $1 AND skill_id = $2
	`, volunteerID, skillID).Scan(&currentWeight)
	
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Volunteer skill not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current weight"})
		return
	}

	// Update weight
	_, err = h.db.Exec(`
		UPDATE volunteer_skills 
		SET skill_weight = $1, updated_at = CURRENT_TIMESTAMP
		WHERE volunteer_id = $2 AND skill_id = $3
	`, req.NewWeight, volunteerID, skillID)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update weight"})
		return
	}

	// Record override history
	_, err = h.db.Exec(`
		INSERT INTO volunteer_skill_weight_overrides 
		(volunteer_id, skill_id, original_weight, override_weight, 
		 adjusted_by_admin_id, initiative_id, reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, volunteerID, skillID, currentWeight, req.NewWeight, 
	   adminUUID, req.InitiativeID, req.Reason)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record weight adjustment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Weight adjusted successfully",
		"volunteer_id": volunteerID,
		"skill_id": skillID,
		"original_weight": currentWeight,
		"new_weight": req.NewWeight,
	})
}

// GetWeightAdjustmentHistory handles GET /api/admin/volunteers/:volunteer_id/weight-history
func (h *AdminVolunteerHandler) GetWeightAdjustmentHistory(c *gin.Context) {
	volunteerIDStr := c.Param("volunteer_id")
	volunteerID, err := uuid.Parse(volunteerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid volunteer ID"})
		return
	}

	query := `
		SELECT 
			wo.id, wo.skill_id, wo.original_weight, wo.override_weight,
			wo.reason, wo.initiative_id, wo.created_at,
			st.skill_name, a.name as admin_name
		FROM volunteer_skill_weight_overrides wo
		JOIN skill_taxonomy st ON wo.skill_id = st.id
		JOIN admins a ON wo.adjusted_by_admin_id = a.id
		WHERE wo.volunteer_id = $1
		ORDER BY wo.created_at DESC
	`

	rows, err := h.db.Query(query, volunteerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get weight history"})
		return
	}
	defer rows.Close()

	var history []gin.H
	for rows.Next() {
		var record struct {
			ID            uuid.UUID  `json:"id"`
			SkillID       int        `json:"skill_id"`
			SkillName     string     `json:"skill_name"`
			OriginalWeight float64   `json:"original_weight"`
			OverrideWeight float64   `json:"override_weight"`
			Reason        *string    `json:"reason"`
			InitiativeID  *uuid.UUID `json:"initiative_id"`
			AdminName     string     `json:"admin_name"`
			CreatedAt     time.Time  `json:"created_at"`
		}

		err := rows.Scan(
			&record.ID, &record.SkillID, &record.OriginalWeight, &record.OverrideWeight,
			&record.Reason, &record.InitiativeID, &record.CreatedAt,
			&record.SkillName, &record.AdminName,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan weight history"})
			return
		}

		history = append(history, gin.H{
			"id": record.ID,
			"skill": gin.H{
				"id": record.SkillID,
				"name": record.SkillName,
			},
			"original_weight": record.OriginalWeight,
			"override_weight": record.OverrideWeight,
			"reason": record.Reason,
			"initiative_id": record.InitiativeID,
			"adjusted_by": record.AdminName,
			"created_at": record.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"volunteer_id": volunteerID,
		"history": history,
		"count": len(history),
	})
}

// GetVolunteerSkills handles GET /api/admin/volunteers/:volunteer_id/skills
func (h *AdminVolunteerHandler) GetVolunteerSkills(c *gin.Context) {
	volunteerIDStr := c.Param("volunteer_id")
	volunteerID, err := uuid.Parse(volunteerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid volunteer ID"})
		return
	}

	skills, err := h.taxonomyService.GetVolunteerSkills(volunteerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get volunteer skills"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"volunteer_id": volunteerID,
		"skills": skills,
		"count": len(skills),
	})
}
