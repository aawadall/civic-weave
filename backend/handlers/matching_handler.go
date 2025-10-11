package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"civicweave/backend/models"
	"civicweave/backend/services"
)

// SkillMatchingHandler handles volunteer-initiative matching endpoints using sparse vectors
type SkillMatchingHandler struct {
	db                *sql.DB
	taxonomyService   *models.SkillTaxonomyService
	matchingService   *services.SkillMatchingService
}

// NewSkillMatchingHandler creates a new skill matching handler
func NewSkillMatchingHandler(db *sql.DB, taxonomyService *models.SkillTaxonomyService, matchingService *services.SkillMatchingService) *SkillMatchingHandler {
	return &SkillMatchingHandler{
		db:              db,
		taxonomyService: taxonomyService,
		matchingService: matchingService,
	}
}

// GetCandidateVolunteers handles GET /api/initiatives/:id/candidate-volunteers
func (h *SkillMatchingHandler) GetCandidateVolunteers(c *gin.Context) {
	initiativeIDStr := c.Param("id")
	initiativeID, err := uuid.Parse(initiativeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid initiative ID"})
		return
	}

	// Get query parameters
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	minScoreStr := c.DefaultQuery("min_score", "0.1")
	minScore, err := strconv.ParseFloat(minScoreStr, 64)
	if err != nil {
		minScore = 0.1
	}

	// Query pre-calculated matches (super fast!)
	query := `
		SELECT 
			v.id, v.name, v.phone, v.location_address,
			m.match_score, m.jaccard_index, 
			m.matched_skill_ids, m.matched_skill_count,
			m.calculated_at
		FROM volunteer_initiative_matches m
		JOIN volunteers v ON m.volunteer_id = v.id
		WHERE m.initiative_id = $1 
			AND m.match_score >= $2
			AND v.skills_visible = true
		ORDER BY m.match_score DESC, m.matched_skill_count DESC
		LIMIT $3
	`

	rows, err := h.db.Query(query, initiativeID, minScore, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get candidate volunteers"})
		return
	}
	defer rows.Close()

	var candidates []gin.H
	for rows.Next() {
		var volunteer struct {
			ID                uuid.UUID `json:"id"`
			Name              string    `json:"name"`
			Phone             string    `json:"phone"`
			LocationAddress   string    `json:"location_address"`
			MatchScore        float64   `json:"match_score"`
			JaccardIndex      float64   `json:"jaccard_index"`
			MatchedSkillIDs   []int     `json:"matched_skill_ids"`
			MatchedSkillCount int       `json:"matched_skill_count"`
			CalculatedAt      time.Time `json:"calculated_at"`
		}

		err := rows.Scan(
			&volunteer.ID, &volunteer.Name, &volunteer.Phone, &volunteer.LocationAddress,
			&volunteer.MatchScore, &volunteer.JaccardIndex,
			&volunteer.MatchedSkillIDs, &volunteer.MatchedSkillCount,
			&volunteer.CalculatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan volunteer data"})
			return
		}

		candidates = append(candidates, gin.H{
			"volunteer": volunteer,
			"match_percentage": int(volunteer.MatchScore * 100),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"candidates": candidates,
		"count":      len(candidates),
		"initiative_id": initiativeID,
		"min_score": minScore,
	})
}

// GetRecommendedInitiatives handles GET /api/volunteers/me/recommended-initiatives
func (h *SkillMatchingHandler) GetRecommendedInitiatives(c *gin.Context) {
	volunteerID, exists := c.Get("volunteer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Volunteer ID not found"})
		return
	}

	volunteerUUID, ok := volunteerID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid volunteer ID"})
		return
	}

	// Get query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 50 {
		limit = 20
	}

	minScoreStr := c.DefaultQuery("min_score", "0.2")
	minScore, err := strconv.ParseFloat(minScoreStr, 64)
	if err != nil {
		minScore = 0.2
	}

	// Query pre-calculated matches
	query := `
		SELECT 
			i.id, i.title, i.description, i.location_address,
			i.start_date, i.end_date, i.status,
			m.match_score, m.matched_skill_count,
			m.matched_skill_ids, m.calculated_at
		FROM volunteer_initiative_matches m
		JOIN initiatives i ON m.initiative_id = i.id
		WHERE m.volunteer_id = $1 
			AND m.match_score >= $2
			AND i.status = 'active'
		ORDER BY m.match_score DESC, m.matched_skill_count DESC
		LIMIT $3
	`

	rows, err := h.db.Query(query, volunteerUUID, minScore, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recommended initiatives"})
		return
	}
	defer rows.Close()

	var recommendations []gin.H
	for rows.Next() {
		var initiative struct {
			ID                uuid.UUID  `json:"id"`
			Title             string     `json:"title"`
			Description       string     `json:"description"`
			LocationAddress   string     `json:"location_address"`
			StartDate         *time.Time `json:"start_date"`
			EndDate           *time.Time `json:"end_date"`
			Status            string     `json:"status"`
			MatchScore        float64    `json:"match_score"`
			MatchedSkillCount int        `json:"matched_skill_count"`
			MatchedSkillIDs   []int      `json:"matched_skill_ids"`
			CalculatedAt      time.Time  `json:"calculated_at"`
		}

		err := rows.Scan(
			&initiative.ID, &initiative.Title, &initiative.Description, &initiative.LocationAddress,
			&initiative.StartDate, &initiative.EndDate, &initiative.Status,
			&initiative.MatchScore, &initiative.MatchedSkillCount,
			&initiative.MatchedSkillIDs, &initiative.CalculatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan initiative data"})
			return
		}

		recommendations = append(recommendations, gin.H{
			"initiative": initiative,
			"match_percentage": int(initiative.MatchScore * 100),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": recommendations,
		"count":          len(recommendations),
		"volunteer_id":   volunteerUUID,
		"min_score":      minScore,
	})
}

// GetMatchExplanation handles GET /api/matching/explanation/:volunteerId/:initiativeId
func (h *SkillMatchingHandler) GetMatchExplanation(c *gin.Context) {
	volunteerIDStr := c.Param("volunteerId")
	volunteerID, err := uuid.Parse(volunteerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid volunteer ID"})
		return
	}

	initiativeIDStr := c.Param("initiativeId")
	initiativeID, err := uuid.Parse(initiativeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid initiative ID"})
		return
	}

	// Get match details
	var match struct {
		MatchScore        float64   `json:"match_score"`
		JaccardIndex      float64   `json:"jaccard_index"`
		MatchedSkillIDs   []int     `json:"matched_skill_ids"`
		MatchedSkillCount int       `json:"matched_skill_count"`
		CalculatedAt      time.Time `json:"calculated_at"`
	}

	err = h.db.QueryRow(`
		SELECT match_score, jaccard_index, matched_skill_ids, matched_skill_count, calculated_at
		FROM volunteer_initiative_matches
		WHERE volunteer_id = $1 AND initiative_id = $2
	`, volunteerID, initiativeID).Scan(
		&match.MatchScore, &match.JaccardIndex,
		&match.MatchedSkillIDs, &match.MatchedSkillCount,
		&match.CalculatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
		return
	}

	// Get skill names for matched skills
	skillNames := make([]string, 0, len(match.MatchedSkillIDs))
	if len(match.MatchedSkillIDs) > 0 {
		skillQuery := `
			SELECT skill_name 
			FROM skill_taxonomy 
			WHERE id = ANY($1)
			ORDER BY skill_name
		`
		rows, err := h.db.Query(skillQuery, match.MatchedSkillIDs)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var skillName string
				if err := rows.Scan(&skillName); err == nil {
					skillNames = append(skillNames, skillName)
				}
			}
		}
	}

	// Get initiative required skills
	initiativeSkills, err := h.taxonomyService.GetInitiativeSkills(initiativeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get initiative skills"})
		return
	}

	// Get volunteer skills
	volunteerSkills, err := h.taxonomyService.GetVolunteerSkills(volunteerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get volunteer skills"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"match_details": match,
		"matched_skills": skillNames,
		"initiative_skills": initiativeSkills,
		"volunteer_skills": volunteerSkills,
		"explanation": gin.H{
			"match_percentage": int(match.MatchScore * 100),
			"skills_matched": match.MatchedSkillCount,
			"total_required": len(initiativeSkills),
			"coverage_percentage": int(float64(match.MatchedSkillCount) / float64(len(initiativeSkills)) * 100),
		},
	})
}

// GetMyMatches handles GET /api/volunteers/me/matches
func (h *SkillMatchingHandler) GetMyMatches(c *gin.Context) {
	// This is an alias for GetRecommendedInitiatives for backward compatibility
	h.GetRecommendedInitiatives(c)
}
