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

// ApplicationHandler handles application-related requests
type ApplicationHandler struct {
	service *models.ApplicationService
	config  *config.Config
}

// NewApplicationHandler creates a new application handler
func NewApplicationHandler(service *models.ApplicationService, config *config.Config) *ApplicationHandler {
	return &ApplicationHandler{
		service: service,
		config:  config,
	}
}

// ListApplications handles GET /api/applications
func (h *ApplicationHandler) ListApplications(c *gin.Context) {
	// Get query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	status := c.Query("status")
	volunteerIDStr := c.Query("volunteer_id")
	initiativeIDStr := c.Query("initiative_id")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Parse optional UUIDs
	var volunteerID, initiativeID *uuid.UUID
	if volunteerIDStr != "" {
		if id, err := uuid.Parse(volunteerIDStr); err == nil {
			volunteerID = &id
		}
	}
	if initiativeIDStr != "" {
		if id, err := uuid.Parse(initiativeIDStr); err == nil {
			initiativeID = &id
		}
	}

	// Get applications
	applications, err := h.service.List(limit, offset, volunteerID, initiativeID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get applications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"applications": applications,
		"limit":        limit,
		"offset":       offset,
		"count":        len(applications),
	})
}

// CreateApplicationRequest represents application creation request
type CreateApplicationRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
	Message   string `json:"message"`
}

// CreateApplication handles POST /api/applications
func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
	var req CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from JWT context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user is a volunteer
	if !userCtx.HasRole("volunteer") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only volunteers can apply to projects"})
		return
	}

	// Get volunteer ID from user ID
	volunteerService := models.NewVolunteerService(h.service.GetDB())
	volunteer, err := volunteerService.GetByUserID(userCtx.ID)
	if err != nil || volunteer == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Volunteer profile not found"})
		return
	}

	// Parse project ID
	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Check if application already exists
	existing, err := h.service.GetByProjectAndVolunteer(projectID, volunteer.ID)
	if err == nil && existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Application already exists"})
		return
	}

	application := &models.Application{
		ProjectID:   projectID,
		VolunteerID: volunteer.ID,
		Status:      "pending",
		AdminNotes:  req.Message, // Store volunteer message in admin_notes for now
	}

	if err := h.service.Create(application); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create application"})
		return
	}

	c.JSON(http.StatusCreated, application)
}

// GetApplication handles GET /api/applications/:id
func (h *ApplicationHandler) GetApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID"})
		return
	}

	application, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get application"})
		return
	}

	if application == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	c.JSON(http.StatusOK, application)
}

// UpdateApplication handles PUT /api/applications/:id
func (h *ApplicationHandler) UpdateApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID"})
		return
	}

	var application models.Application
	if err := c.ShouldBindJSON(&application); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	application.ID = id

	if err := h.service.Update(&application); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update application"})
		return
	}

	c.JSON(http.StatusOK, application)
}

// DeleteApplication handles DELETE /api/applications/:id
func (h *ApplicationHandler) DeleteApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete application"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
