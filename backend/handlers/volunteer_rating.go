package handlers

import (
	"net/http"
	"strconv"

	"civicweave/backend/config"
	"civicweave/backend/middleware"
	"civicweave/backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// VolunteerRatingHandler handles volunteer rating requests
type VolunteerRatingHandler struct {
	ratingService    *models.VolunteerRatingService
	volunteerService *models.VolunteerService
	config           *config.Config
}

// NewVolunteerRatingHandler creates a new volunteer rating handler
func NewVolunteerRatingHandler(ratingService *models.VolunteerRatingService, volunteerService *models.VolunteerService, config *config.Config) *VolunteerRatingHandler {
	return &VolunteerRatingHandler{
		ratingService:    ratingService,
		volunteerService: volunteerService,
		config:           config,
	}
}

// CreateRating handles POST /api/volunteers/:id/ratings
func (h *VolunteerRatingHandler) CreateRating(c *gin.Context) {
	volunteerIDStr := c.Param("id")
	volunteerID, err := uuid.Parse(volunteerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid volunteer ID"})
		return
	}

	var req struct {
		SkillClaimID *uuid.UUID        `json:"skill_claim_id"`
		ProjectID    *uuid.UUID        `json:"project_id"`
		Rating       models.RatingType `json:"rating" binding:"required"`
		Notes        string            `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate rating type
	if req.Rating != models.RatingUp && req.Rating != models.RatingDown && req.Rating != models.RatingNeutral {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating type"})
		return
	}

	// Get rater user ID from context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has permission to rate (team_lead or admin)
	if !userCtx.HasAnyRole("team_lead", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to rate volunteers"})
		return
	}

	// Check if user has already rated this volunteer for this skill/project
	hasRated, err := h.ratingService.HasRated(userCtx.ID, volunteerID, req.SkillClaimID, req.ProjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing rating"})
		return
	}
	if hasRated {
		c.JSON(http.StatusConflict, gin.H{"error": "You have already rated this volunteer for this skill/project"})
		return
	}

	rating := &models.VolunteerRating{
		VolunteerID:   volunteerID,
		SkillClaimID:  req.SkillClaimID,
		RatedByUserID: userCtx.ID,
		ProjectID:     req.ProjectID,
		Rating:        req.Rating,
		Notes:         req.Notes,
	}

	if err := h.ratingService.CreateRating(rating); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rating"})
		return
	}

	c.JSON(http.StatusCreated, rating)
}

// GetVolunteerScorecard handles GET /api/volunteers/:id/scorecard
func (h *VolunteerRatingHandler) GetVolunteerScorecard(c *gin.Context) {
	volunteerIDStr := c.Param("id")
	volunteerID, err := uuid.Parse(volunteerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid volunteer ID"})
		return
	}

	// Get user context to check permissions
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has permission to view scorecard (team_lead, admin, or the volunteer themselves)
	if !userCtx.HasAnyRole("team_lead", "admin") {
		// Check if the user is viewing their own scorecard
		volunteer, err := h.volunteerService.GetByID(volunteerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get volunteer"})
			return
		}
		if volunteer == nil || volunteer.UserID != userCtx.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to view this scorecard"})
			return
		}
	}

	scorecard, err := h.ratingService.GetVolunteerScorecard(volunteerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get volunteer scorecard"})
		return
	}

	c.JSON(http.StatusOK, scorecard)
}

// ListRatingsForVolunteer handles GET /api/volunteers/:id/ratings
func (h *VolunteerRatingHandler) ListRatingsForVolunteer(c *gin.Context) {
	volunteerIDStr := c.Param("id")
	volunteerID, err := uuid.Parse(volunteerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid volunteer ID"})
		return
	}

	// Get query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get user context to check permissions
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has permission to view ratings (team_lead, admin, or the volunteer themselves)
	if !userCtx.HasAnyRole("team_lead", "admin") {
		// Check if the user is viewing their own ratings
		volunteer, err := h.volunteerService.GetByID(volunteerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get volunteer"})
			return
		}
		if volunteer == nil || volunteer.UserID != userCtx.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to view these ratings"})
			return
		}
	}

	ratings, err := h.ratingService.ListRatingsForVolunteer(volunteerID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get volunteer ratings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ratings": ratings,
		"limit":   limit,
		"offset":  offset,
		"count":   len(ratings),
	})
}

// ListRatingsByRater handles GET /api/ratings/my-ratings
func (h *VolunteerRatingHandler) ListRatingsByRater(c *gin.Context) {
	// Get query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has permission to view their ratings (team_lead or admin)
	if !userCtx.HasAnyRole("team_lead", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to view ratings"})
		return
	}

	ratings, err := h.ratingService.ListRatingsByRater(userCtx.ID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ratings by rater"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ratings": ratings,
		"limit":   limit,
		"offset":  offset,
		"count":   len(ratings),
	})
}

// UpdateRating handles PUT /api/ratings/:id
func (h *VolunteerRatingHandler) UpdateRating(c *gin.Context) {
	ratingIDStr := c.Param("id")
	ratingID, err := uuid.Parse(ratingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating ID"})
		return
	}

	var req struct {
		Rating models.RatingType `json:"rating" binding:"required"`
		Notes  string            `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate rating type
	if req.Rating != models.RatingUp && req.Rating != models.RatingDown && req.Rating != models.RatingNeutral {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating type"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get the rating to check ownership
	rating, err := h.ratingService.GetRatingByID(ratingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rating"})
		return
	}
	if rating == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rating not found"})
		return
	}

	// Check if user has permission to update this rating (admin or the original rater)
	if !userCtx.HasRole("admin") && rating.RatedByUserID != userCtx.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to update this rating"})
		return
	}

	rating.Rating = req.Rating
	rating.Notes = req.Notes

	if err := h.ratingService.UpdateRating(rating); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rating"})
		return
	}

	c.JSON(http.StatusOK, rating)
}

// DeleteRating handles DELETE /api/ratings/:id
func (h *VolunteerRatingHandler) DeleteRating(c *gin.Context) {
	ratingIDStr := c.Param("id")
	ratingID, err := uuid.Parse(ratingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get the rating to check ownership
	rating, err := h.ratingService.GetRatingByID(ratingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rating"})
		return
	}
	if rating == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rating not found"})
		return
	}

	// Check if user has permission to delete this rating (admin or the original rater)
	if !userCtx.HasRole("admin") && rating.RatedByUserID != userCtx.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to delete this rating"})
		return
	}

	if err := h.ratingService.DeleteRating(ratingID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rating"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetTopRatedVolunteers handles GET /api/volunteers/top-rated
func (h *VolunteerRatingHandler) GetTopRatedVolunteers(c *gin.Context) {
	// Get query parameters
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	// Get user context to check permissions
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has permission to view top-rated volunteers (team_lead or admin)
	if !userCtx.HasAnyRole("team_lead", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to view top-rated volunteers"})
		return
	}

	volunteers, err := h.ratingService.GetTopRatedVolunteers(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get top-rated volunteers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"volunteers": volunteers,
		"count":      len(volunteers),
		"limit":      limit,
	})
}
