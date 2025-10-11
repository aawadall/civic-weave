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

// ProjectHandler handles project-related requests
type ProjectHandler struct {
	service          *models.ProjectService
	geocodingService *utils.GeocodingService
	config           *config.Config
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(service *models.ProjectService, geocodingService *utils.GeocodingService, config *config.Config) *ProjectHandler {
	return &ProjectHandler{
		service:          service,
		geocodingService: geocodingService,
		config:           config,
	}
}

// ListProjects handles GET /api/projects
func (h *ProjectHandler) ListProjects(c *gin.Context) {
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

	// Get projects
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}
	projects, err := h.service.List(limit, offset, statusPtr, skillsParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get projects"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"limit":    limit,
		"offset":   offset,
		"count":    len(projects),
	})
}

// CreateProjectRequest represents project creation request
type CreateProjectRequest struct {
	Title           string   `json:"title" binding:"required"`
	Description     string   `json:"description" binding:"required"`
	RequiredSkills  []string `json:"required_skills"`
	LocationAddress string   `json:"location_address"`
	StartDate       string   `json:"start_date"`
	EndDate         string   `json:"end_date"`
	Status          string   `json:"status"`
	ProjectStatus   string   `json:"project_status"`
	TeamLeadID      *string  `json:"team_lead_id"`
}

// CreateProject handles POST /api/projects
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req CreateProjectRequest
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

	// Check if user has permission to create projects (team_lead or admin)
	if !userCtx.HasAnyRole("team_lead", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to create projects"})
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

	// Set default project status
	projectStatus := models.ProjectStatusDraft
	if req.ProjectStatus != "" {
		projectStatus = models.ProjectStatus(req.ProjectStatus)
	}

	// Parse team lead ID if provided
	var teamLeadID *uuid.UUID
	if req.TeamLeadID != nil && *req.TeamLeadID != "" {
		if parsed, err := uuid.Parse(*req.TeamLeadID); err == nil {
			teamLeadID = &parsed
		}
	}

	project := &models.Project{
		Title:            req.Title,
		Description:      req.Description,
		RequiredSkills:   req.RequiredSkills,
		LocationLat:      locationLat,
		LocationLng:      locationLng,
		LocationAddress:  req.LocationAddress,
		StartDate:        startDate,
		EndDate:          endDate,
		Status:           status,
		ProjectStatus:    projectStatus,
		CreatedByAdminID: userCtx.ID,
		TeamLeadID:       teamLeadID,
	}

	if err := h.service.Create(project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// GetProject handles GET /api/projects/:id
func (h *ProjectHandler) GetProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	project, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project"})
		return
	}

	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// UpdateProject handles PUT /api/projects/:id
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project.ID = id

	if err := h.service.Update(&project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject handles DELETE /api/projects/:id
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetProjectWithDetails handles GET /api/projects/:id/details
func (h *ProjectHandler) GetProjectWithDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	project, err := h.service.GetByIDWithDetails(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project details"})
		return
	}

	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// GetProjectSignups handles GET /api/projects/:id/signups
func (h *ProjectHandler) GetProjectSignups(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has permission to view signups (team_lead or admin)
	if !userCtx.HasAnyRole("team_lead", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to view project signups"})
		return
	}

	// Check if user is team lead for this project or admin
	isTeamLead, err := h.service.IsTeamLead(id, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team lead status"})
		return
	}

	if !userCtx.HasRole("admin") && !isTeamLead {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to view signups for this project"})
		return
	}

	signups, err := h.service.GetProjectSignups(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project signups"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"signups": signups})
}

// GetProjectTeamMembers handles GET /api/projects/:id/team-members
func (h *ProjectHandler) GetProjectTeamMembers(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has permission to view team members (team_lead or admin)
	if !userCtx.HasAnyRole("team_lead", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to view team members"})
		return
	}

	// Check if user is team lead for this project or admin
	isTeamLead, err := h.service.IsTeamLead(id, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team lead status"})
		return
	}

	if !userCtx.HasRole("admin") && !isTeamLead {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to view team members for this project"})
		return
	}

	teamMembers, err := h.service.GetProjectTeamMembers(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get team members"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"team_members": teamMembers})
}

// AddTeamMember handles POST /api/projects/:id/team-members
func (h *ProjectHandler) AddTeamMember(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req struct {
		VolunteerID uuid.UUID `json:"volunteer_id" binding:"required"`
		Status      string    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has permission to add team members (team_lead or admin)
	if !userCtx.HasAnyRole("team_lead", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to add team members"})
		return
	}

	// Check if user is team lead for this project or admin
	isTeamLead, err := h.service.IsTeamLead(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team lead status"})
		return
	}

	if !userCtx.HasRole("admin") && !isTeamLead {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to add team members to this project"})
		return
	}

	// Set default status
	status := models.TeamMemberStatusInvited
	if req.Status != "" {
		status = models.TeamMemberStatus(req.Status)
	}

	if err := h.service.AddTeamMember(projectID, req.VolunteerID, status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add team member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team member added successfully"})
}

// UpdateTeamMemberStatus handles PUT /api/projects/:id/team-members/:volunteerId
func (h *ProjectHandler) UpdateTeamMemberStatus(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	volunteerIDStr := c.Param("volunteerId")
	volunteerID, err := uuid.Parse(volunteerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid volunteer ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has permission to update team member status (team_lead or admin)
	if !userCtx.HasAnyRole("team_lead", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to update team member status"})
		return
	}

	// Check if user is team lead for this project or admin
	isTeamLead, err := h.service.IsTeamLead(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team lead status"})
		return
	}

	if !userCtx.HasRole("admin") && !isTeamLead {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to update team member status for this project"})
		return
	}

	status := models.TeamMemberStatus(req.Status)
	if err := h.service.UpdateTeamMemberStatus(projectID, volunteerID, status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team member status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team member status updated successfully"})
}

// AssignTeamLead handles PUT /api/projects/:id/team-lead
func (h *ProjectHandler) AssignTeamLead(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req struct {
		TeamLeadID uuid.UUID `json:"team_lead_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has permission to assign team lead (admin only)
	if !userCtx.HasRole("admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to assign team lead"})
		return
	}

	if err := h.service.AssignTeamLead(projectID, req.TeamLeadID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign team lead"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team lead assigned successfully"})
}
