package handlers

import (
	"net/http"
	"strconv"
	"time"

	"civicweave/backend/config"
	"civicweave/backend/middleware"
	"civicweave/backend/models"
	"civicweave/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// InitiativeHandler handles initiative-related requests
type InitiativeHandler struct {
	service          *models.InitiativeService
	geocodingService *utils.GeocodingService
	config           *config.Config
}

// NewInitiativeHandler creates a new initiative handler
func NewInitiativeHandler(service *models.InitiativeService, geocodingService *utils.GeocodingService, config *config.Config) *InitiativeHandler {
	return &InitiativeHandler{
		service:          service,
		geocodingService: geocodingService,
		config:           config,
	}
}

// ListInitiatives handles GET /api/initiatives
func (h *InitiativeHandler) ListInitiatives(c *gin.Context) {
	// Get query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	status := c.Query("status")
	skillsParam := c.QueryArray("skills")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get initiatives
	initiatives, err := h.service.List(limit, offset, status, skillsParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get initiatives"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"initiatives": initiatives,
		"limit":       limit,
		"offset":      offset,
		"count":       len(initiatives),
	})
}

// CreateInitiativeRequest represents initiative creation request
type CreateInitiativeRequest struct {
	Title           string   `json:"title" binding:"required"`
	Description     string   `json:"description" binding:"required"`
	RequiredSkills  []string `json:"required_skills"`
	LocationAddress string   `json:"location_address"`
	StartDate       string   `json:"start_date"`
	EndDate         string   `json:"end_date"`
	Status          string   `json:"status"`
}

// CreateInitiative handles POST /api/initiatives
func (h *InitiativeHandler) CreateInitiative(c *gin.Context) {
	var req CreateInitiativeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get admin ID from JWT context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Geocode location if provided
	var locationLat, locationLng *float64
	if req.LocationAddress != "" {
		lat, lng, _, err := h.geocodingService.GeocodeAddress(req.LocationAddress)
		if err != nil {
			// Log error but don't fail creation
			// Location can be added later
		} else {
			locationLat = &lat
			locationLng = &lng
		}
	}

	// Parse dates
	var startDate, endDate *time.Time
	if req.StartDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = &parsed
		}
	}
	if req.EndDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			endDate = &parsed
		}
	}

	// Set default status
	status := req.Status
	if status == "" {
		status = "draft"
	}

	initiative := &models.Initiative{
		Title:            req.Title,
		Description:      req.Description,
		RequiredSkills:   req.RequiredSkills,
		LocationLat:      locationLat,
		LocationLng:      locationLng,
		LocationAddress:  req.LocationAddress,
		StartDate:        startDate,
		EndDate:          endDate,
		Status:           status,
		CreatedByAdminID: userCtx.ID, // This should be the admin ID, not user ID
	}

	if err := h.service.Create(initiative); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create initiative"})
		return
	}

	c.JSON(http.StatusCreated, initiative)
}

// GetInitiative handles GET /api/initiatives/:id
func (h *InitiativeHandler) GetInitiative(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid initiative ID"})
		return
	}

	initiative, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get initiative"})
		return
	}

	if initiative == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Initiative not found"})
		return
	}

	c.JSON(http.StatusOK, initiative)
}

// UpdateInitiative handles PUT /api/initiatives/:id
func (h *InitiativeHandler) UpdateInitiative(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid initiative ID"})
		return
	}

	var initiative models.Initiative
	if err := c.ShouldBindJSON(&initiative); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	initiative.ID = id

	if err := h.service.Update(&initiative); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update initiative"})
		return
	}

	c.JSON(http.StatusOK, initiative)
}

// DeleteInitiative handles DELETE /api/initiatives/:id
func (h *InitiativeHandler) DeleteInitiative(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid initiative ID"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete initiative"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
