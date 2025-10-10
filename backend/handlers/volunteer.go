package handlers

import (
	"net/http"
	"strconv"

	"civicweave/backend/config"
	"civicweave/backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// VolunteerHandler handles volunteer-related requests
type VolunteerHandler struct {
	service *models.VolunteerService
	config  *config.Config
}

// NewVolunteerHandler creates a new volunteer handler
func NewVolunteerHandler(service *models.VolunteerService, config *config.Config) *VolunteerHandler {
	return &VolunteerHandler{
		service: service,
		config:  config,
	}
}

// ListVolunteers handles GET /api/volunteers
func (h *VolunteerHandler) ListVolunteers(c *gin.Context) {
	// Get query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	skillsParam := c.QueryArray("skills")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get volunteers
	volunteers, err := h.service.List(limit, offset, skillsParam, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get volunteers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"volunteers": volunteers,
		"limit":      limit,
		"offset":     offset,
		"count":      len(volunteers),
	})
}

// CreateVolunteer handles POST /api/volunteers
func (h *VolunteerHandler) CreateVolunteer(c *gin.Context) {
	var volunteer models.Volunteer
	if err := c.ShouldBindJSON(&volunteer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Create(&volunteer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create volunteer"})
		return
	}

	c.JSON(http.StatusCreated, volunteer)
}

// GetVolunteer handles GET /api/volunteers/:id
func (h *VolunteerHandler) GetVolunteer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid volunteer ID"})
		return
	}

	volunteer, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get volunteer"})
		return
	}

	if volunteer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Volunteer not found"})
		return
	}

	c.JSON(http.StatusOK, volunteer)
}

// UpdateVolunteer handles PUT /api/volunteers/:id
func (h *VolunteerHandler) UpdateVolunteer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid volunteer ID"})
		return
	}

	var volunteer models.Volunteer
	if err := c.ShouldBindJSON(&volunteer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	volunteer.ID = id

	if err := h.service.Update(&volunteer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update volunteer"})
		return
	}

	c.JSON(http.StatusOK, volunteer)
}
