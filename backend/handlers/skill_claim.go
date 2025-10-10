package handlers

import (
	"net/http"
	"strconv"

	"civicweave/backend/config"
	"civicweave/backend/models"
	"civicweave/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SkillClaimHandler handles skill claim-related requests
type SkillClaimHandler struct {
	skillClaimService        *models.SkillClaimService
	vectorAggregationService *services.VectorAggregationService
	vectorMatchingService    *services.VectorMatchingService
	embeddingService         *services.EmbeddingService
	config                   *config.Config
}

// NewSkillClaimHandler creates a new skill claim handler
func NewSkillClaimHandler(
	skillClaimService *models.SkillClaimService,
	vectorAggregationService *services.VectorAggregationService,
	vectorMatchingService *services.VectorMatchingService,
	embeddingService *services.EmbeddingService,
	config *config.Config,
) *SkillClaimHandler {
	return &SkillClaimHandler{
		skillClaimService:        skillClaimService,
		vectorAggregationService: vectorAggregationService,
		vectorMatchingService:    vectorMatchingService,
		embeddingService:         embeddingService,
		config:                   config,
	}
}

// CreateSkillClaimRequest represents the request to create a skill claim
type CreateSkillClaimRequest struct {
	ClaimText string `json:"claim_text" binding:"required,min=10,max=500"`
}

// UpdateSkillWeightRequest represents the request to update a skill weight
type UpdateSkillWeightRequest struct {
	Weight float64 `json:"weight" binding:"required,min=0.1,max=1.0"`
	Reason string  `json:"reason,omitempty"`
}

// UpdateSkillsVisibilityRequest represents the request to update skills visibility
type UpdateSkillsVisibilityRequest struct {
	Visible bool `json:"visible"`
}

// CreateSkillClaim handles POST /api/volunteers/me/skills
func (h *SkillClaimHandler) CreateSkillClaim(c *gin.Context) {
	// Get volunteer ID from context (set by auth middleware)
	volunteerIDInterface, exists := c.Get("volunteer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Volunteer ID not found in context"})
		return
	}

	volunteerID, ok := volunteerIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid volunteer ID format"})
		return
	}

	var req CreateSkillClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate embedding for the claim text
	embedding, err := h.embeddingService.GenerateEmbedding(req.ClaimText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate skill embedding"})
		return
	}

	// Create skill claim
	claim, err := h.skillClaimService.CreateSkillClaim(volunteerID, req.ClaimText, embedding)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create skill claim"})
		return
	}

	// Trigger vector aggregation
	err = h.vectorAggregationService.TriggerAggregationOnClaimChange(volunteerID)
	if err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Skill claim created successfully",
		"claim":   claim,
	})
}

// GetMySkillClaims handles GET /api/volunteers/me/skills
func (h *SkillClaimHandler) GetMySkillClaims(c *gin.Context) {
	// Get volunteer ID from context
	volunteerIDInterface, exists := c.Get("volunteer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Volunteer ID not found in context"})
		return
	}

	volunteerID, ok := volunteerIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid volunteer ID format"})
		return
	}

	// Get skill claims with weights
	claims, err := h.skillClaimService.GetActiveClaimsByVolunteer(volunteerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get skill claims"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"claims": claims,
		"count":  len(claims),
	})
}

// DeleteSkillClaim handles DELETE /api/volunteers/me/skills/:id
func (h *SkillClaimHandler) DeleteSkillClaim(c *gin.Context) {
	// Get volunteer ID from context
	volunteerIDInterface, exists := c.Get("volunteer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Volunteer ID not found in context"})
		return
	}

	volunteerID, ok := volunteerIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid volunteer ID format"})
		return
	}

	// Get claim ID from URL
	claimIDStr := c.Param("id")
	claimID, err := uuid.Parse(claimIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid claim ID"})
		return
	}

	// Verify the claim belongs to the volunteer
	claim, err := h.skillClaimService.GetSkillClaimByID(claimID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get skill claim"})
		return
	}

	if claim == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Skill claim not found"})
		return
	}

	if claim.VolunteerID != volunteerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own skill claims"})
		return
	}

	// Deactivate the claim
	err = h.skillClaimService.DeactivateSkillClaim(claimID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete skill claim"})
		return
	}

	// Trigger vector aggregation
	err = h.vectorAggregationService.TriggerAggregationOnClaimChange(volunteerID)
	if err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	c.JSON(http.StatusOK, gin.H{"message": "Skill claim deleted successfully"})
}

// UpdateSkillWeight handles PATCH /api/admin/skill-claims/:id/weight
func (h *SkillClaimHandler) UpdateSkillWeight(c *gin.Context) {
	// Get admin ID from context
	adminIDInterface, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin ID not found in context"})
		return
	}

	adminID, ok := adminIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid admin ID format"})
		return
	}

	// Get claim ID from URL
	claimIDStr := c.Param("id")
	claimID, err := uuid.Parse(claimIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid claim ID"})
		return
	}

	var req UpdateSkillWeightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default reason if not provided
	if req.Reason == "" {
		req.Reason = "admin_review"
	}

	// Update the skill weight
	err = h.skillClaimService.UpdateSkillWeight(claimID, req.Weight, &adminID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update skill weight"})
		return
	}

	// Get the claim to trigger aggregation for the volunteer
	claim, err := h.skillClaimService.GetSkillClaimByID(claimID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get skill claim"})
		return
	}

	if claim != nil {
		// Trigger vector aggregation
		err = h.vectorAggregationService.TriggerAggregationOnWeightChange(claim.VolunteerID)
		if err != nil {
			// Log error but don't fail the request
			// TODO: Add proper logging
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Skill weight updated successfully"})
}

// UpdateSkillsVisibility handles PUT /api/volunteers/me/skills-visibility
func (h *SkillClaimHandler) UpdateSkillsVisibility(c *gin.Context) {
	// Get volunteer ID from context
	volunteerIDInterface, exists := c.Get("volunteer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Volunteer ID not found in context"})
		return
	}

	volunteerID, ok := volunteerIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid volunteer ID format"})
		return
	}

	var req UpdateSkillsVisibilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update skills visibility
	err := h.skillClaimService.UpdateVolunteerSkillsVisibility(volunteerID, req.Visible)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update skills visibility"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Skills visibility updated successfully",
		"visible": req.Visible,
	})
}

// GetSkillsVisibility handles GET /api/volunteers/me/skills-visibility
func (h *SkillClaimHandler) GetSkillsVisibility(c *gin.Context) {
	// Get volunteer ID from context
	volunteerIDInterface, exists := c.Get("volunteer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Volunteer ID not found in context"})
		return
	}

	volunteerID, ok := volunteerIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid volunteer ID format"})
		return
	}

	// Get skills visibility
	visible, err := h.skillClaimService.GetVolunteerSkillsVisibility(volunteerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get skills visibility"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"visible": visible})
}

// ListAllSkillClaims handles GET /api/admin/skill-claims (admin only)
func (h *SkillClaimHandler) ListAllSkillClaims(c *gin.Context) {
	// Get query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	activeOnlyStr := c.DefaultQuery("active_only", "true")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	activeOnly := activeOnlyStr == "true"

	// Get all skill claims
	claims, err := h.skillClaimService.ListAllSkillClaims(limit, offset, activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get skill claims"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"claims": claims,
		"limit":  limit,
		"offset": offset,
		"count":  len(claims),
	})
}

// GetTopMatches handles GET /api/volunteers/me/matches
func (h *SkillClaimHandler) GetTopMatches(c *gin.Context) {
	// Get volunteer ID from context
	volunteerIDInterface, exists := c.Get("volunteer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Volunteer ID not found in context"})
		return
	}

	volunteerID, ok := volunteerIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid volunteer ID format"})
		return
	}

	// Get query parameters
	kStr := c.DefaultQuery("limit", "5")
	k, err := strconv.Atoi(kStr)
	if err != nil || k < 1 || k > 20 {
		k = 5
	}

	// Get volunteer's vector
	volunteerVector, err := h.vectorAggregationService.GetVolunteerVector(volunteerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get volunteer vector"})
		return
	}

	if volunteerVector == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "No skill vector found. Add some skills to get matches!",
			"matches": []interface{}{},
		})
		return
	}

	// Find similar volunteers to get an idea of potential matches
	similarVolunteers, err := h.vectorMatchingService.FindSimilarVolunteers(volunteerID, k, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find similar volunteers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"matches": similarVolunteers,
		"count":   len(similarVolunteers),
	})
}

// GetMatchExplanation handles GET /api/volunteers/me/matches/:initiative_id/explanation
func (h *SkillClaimHandler) GetMatchExplanation(c *gin.Context) {
	// Get volunteer ID from context
	volunteerIDInterface, exists := c.Get("volunteer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Volunteer ID not found in context"})
		return
	}

	volunteerID, ok := volunteerIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid volunteer ID format"})
		return
	}

	// Get initiative ID from URL
	initiativeIDStr := c.Param("initiative_id")
	initiativeID, err := uuid.Parse(initiativeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid initiative ID"})
		return
	}

	// Get match explanation
	explanation, err := h.vectorMatchingService.GetVolunteerMatchExplanations(volunteerID, initiativeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get match explanation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"explanation":   explanation,
		"volunteer_id":  volunteerID.String(),
		"initiative_id": initiativeID.String(),
	})
}
