package handlers

import (
	"net/http"
	"strconv"

	"civicweave/backend/config"
	"civicweave/backend/middleware"
	"civicweave/backend/models"
	"civicweave/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MatchingHandler handles matching-related requests
type MatchingHandler struct {
	matchingService   *services.MatchingService
	volunteerService  *models.VolunteerService
	initiativeService *models.InitiativeService
	config            *config.Config
}

// NewMatchingHandler creates a new matching handler
func NewMatchingHandler(matchingService *services.MatchingService, volunteerService *models.VolunteerService, initiativeService *models.InitiativeService, config *config.Config) *MatchingHandler {
	return &MatchingHandler{
		matchingService:   matchingService,
		volunteerService:  volunteerService,
		initiativeService: initiativeService,
		config:            config,
	}
}

// GetMatchesForVolunteer handles GET /api/matching/volunteer/:id
func (h *MatchingHandler) GetMatchesForVolunteer(c *gin.Context) {
	volunteerID := c.Param("id")
	if volunteerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Volunteer ID is required"})
		return
	}

	// Get limit from query parameter
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	matches, err := h.matchingService.GetMatchesForVolunteer(volunteerID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get matches"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"matches": matches,
		"count":   len(matches),
		"limit":   limit,
	})
}

// GetMatchesForInitiative handles GET /api/matching/initiative/:id
func (h *MatchingHandler) GetMatchesForInitiative(c *gin.Context) {
	initiativeID := c.Param("id")
	if initiativeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Initiative ID is required"})
		return
	}

	// Get limit from query parameter
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	matches, err := h.matchingService.GetMatchesForInitiative(initiativeID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get matches"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"matches": matches,
		"count":   len(matches),
		"limit":   limit,
	})
}

// GetMyMatches handles GET /api/matching/my-matches
func (h *MatchingHandler) GetMyMatches(c *gin.Context) {
	// Get volunteer ID from JWT context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get limit from query parameter
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	matches, err := h.matchingService.GetMatchesForVolunteer(userCtx.ID.String(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get matches"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"matches": matches,
		"count":   len(matches),
		"limit":   limit,
	})
}

// GetMatchExplanation handles GET /api/matching/explanation/:volunteerId/:initiativeId
func (h *MatchingHandler) GetMatchExplanation(c *gin.Context) {
	volunteerID := c.Param("volunteerId")
	initiativeID := c.Param("initiativeId")

	if volunteerID == "" || initiativeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both volunteer ID and initiative ID are required"})
		return
	}

	// Parse UUIDs
	volunteerUUID, err := uuid.Parse(volunteerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid volunteer ID"})
		return
	}

	initiativeUUID, err := uuid.Parse(initiativeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid initiative ID"})
		return
	}

	// Get volunteer and initiative
	volunteer, err := h.volunteerService.GetByID(volunteerUUID)
	if err != nil || volunteer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Volunteer not found"})
		return
	}

	initiative, err := h.initiativeService.GetByID(initiativeUUID)
	if err != nil || initiative == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Initiative not found"})
		return
	}

	// Calculate scores
	totalScore, skillScore, locationScore := h.matchingService.CalculateMatchScore(volunteer, initiative)
	explanation := h.matchingService.GetMatchingExplanation(volunteer, initiative)

	c.JSON(http.StatusOK, gin.H{
		"total_score":    totalScore,
		"skill_score":    skillScore,
		"location_score": locationScore,
		"explanation":    explanation,
		"volunteer_id":   volunteerID,
		"initiative_id":  initiativeID,
	})
}
