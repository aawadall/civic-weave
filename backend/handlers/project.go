package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"
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

	log.Printf("📋 LIST_PROJECTS: Fetching projects - limit=%d, offset=%d, status=%v, skills=%v", limit, offset, status, skillsParam)

	// Get projects
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}
	projects, err := h.service.List(limit, offset, statusPtr, skillsParam)
	if err != nil {
		log.Printf("❌ LIST_PROJECTS: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get projects", "details": err.Error()})
		return
	}

	log.Printf("✅ LIST_PROJECTS: Successfully fetched %d projects", len(projects))
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
		log.Printf("❌ CREATE_PROJECT: JSON binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("📝 CREATE_PROJECT: Request data - Title=%s, Description length=%d, Skills=%v", req.Title, len(req.Description), req.RequiredSkills)

	// Get user ID from JWT context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		log.Printf("❌ CREATE_PROJECT: User not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	log.Printf("👤 CREATE_PROJECT: User=%s, Roles=%v", userCtx.ID, userCtx.Roles)

	// Check if user has permission to create projects (team_lead or admin)
	if !userCtx.HasAnyRole("team_lead", "admin") {
		log.Printf("❌ CREATE_PROJECT: Insufficient permissions for user %s", userCtx.ID)
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

	log.Printf("💾 CREATE_PROJECT: Attempting to save project to database")
	if err := h.service.Create(project); err != nil {
		log.Printf("❌ CREATE_PROJECT: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project", "details": err.Error()})
		return
	}

	log.Printf("✅ CREATE_PROJECT: Successfully created project ID=%s", project.ID)
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

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user can edit project (team lead, admin, or creator)
	canEdit, err := h.service.CanEditProject(id, userCtx.ID)
	if err != nil {
		log.Printf("❌ UPDATE_PROJECT: Failed to check edit permissions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
		return
	}

	if !canEdit {
		log.Printf("❌ UPDATE_PROJECT: User %s is not authorized to edit project %s", userCtx.ID, id)
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project team lead, admin, or creator can edit this project"})
		return
	}

	// Get current project to check status restrictions
	currentProject, err := h.service.GetByID(id)
	if err != nil {
		log.Printf("❌ UPDATE_PROJECT: Failed to get current project: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current project"})
		return
	}
	if currentProject == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var updateData models.Project
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Apply field restrictions based on current project status
	restrictedProject := h.applyFieldRestrictions(currentProject, &updateData, userCtx.HasRole("admin"))
	restrictedProject.ID = id

	log.Printf("📝 UPDATE_PROJECT: User %s updating project %s (status: %s)", userCtx.ID, id, currentProject.ProjectStatus)
	if err := h.service.Update(restrictedProject); err != nil {
		log.Printf("❌ UPDATE_PROJECT: Failed to update project: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	log.Printf("✅ UPDATE_PROJECT: Successfully updated project %s", id)
	c.JSON(http.StatusOK, restrictedProject)
}

// applyFieldRestrictions applies field-level restrictions based on project status
func (h *ProjectHandler) applyFieldRestrictions(currentProject *models.Project, updateData *models.Project, isAdmin bool) *models.Project {
	// Start with current project as base
	restricted := *currentProject

	// Admin can override all restrictions
	if isAdmin {
		// Admin can update all fields
		restricted.Title = updateData.Title
		restricted.Description = updateData.Description
		restricted.RequiredSkills = updateData.RequiredSkills
		restricted.LocationLat = updateData.LocationLat
		restricted.LocationLng = updateData.LocationLng
		restricted.LocationAddress = updateData.LocationAddress
		restricted.StartDate = updateData.StartDate
		restricted.EndDate = updateData.EndDate
		restricted.TeamLeadID = updateData.TeamLeadID
		restricted.BudgetTotal = updateData.BudgetTotal
		restricted.BudgetSpent = updateData.BudgetSpent
		restricted.AutoNotifyMatches = updateData.AutoNotifyMatches
		return &restricted
	}

	// Apply restrictions based on current status
	switch currentProject.ProjectStatus {
	case models.ProjectStatusDraft:
		// All fields editable in draft
		restricted.Title = updateData.Title
		restricted.Description = updateData.Description
		restricted.RequiredSkills = updateData.RequiredSkills
		restricted.LocationLat = updateData.LocationLat
		restricted.LocationLng = updateData.LocationLng
		restricted.LocationAddress = updateData.LocationAddress
		restricted.StartDate = updateData.StartDate
		restricted.EndDate = updateData.EndDate
		restricted.TeamLeadID = updateData.TeamLeadID
		restricted.BudgetTotal = updateData.BudgetTotal
		restricted.BudgetSpent = updateData.BudgetSpent
		restricted.AutoNotifyMatches = updateData.AutoNotifyMatches

	case models.ProjectStatusRecruiting:
		// Restrict title, description, required_skills
		restricted.LocationLat = updateData.LocationLat
		restricted.LocationLng = updateData.LocationLng
		restricted.LocationAddress = updateData.LocationAddress
		restricted.StartDate = updateData.StartDate
		restricted.EndDate = updateData.EndDate
		restricted.TeamLeadID = updateData.TeamLeadID
		restricted.BudgetTotal = updateData.BudgetTotal
		restricted.BudgetSpent = updateData.BudgetSpent
		restricted.AutoNotifyMatches = updateData.AutoNotifyMatches

	case models.ProjectStatusActive:
		// Restrict title, required_skills, start_date
		restricted.Description = updateData.Description
		restricted.LocationLat = updateData.LocationLat
		restricted.LocationLng = updateData.LocationLng
		restricted.LocationAddress = updateData.LocationAddress
		restricted.EndDate = updateData.EndDate
		restricted.TeamLeadID = updateData.TeamLeadID
		restricted.BudgetTotal = updateData.BudgetTotal
		restricted.BudgetSpent = updateData.BudgetSpent
		restricted.AutoNotifyMatches = updateData.AutoNotifyMatches

	case models.ProjectStatusCompleted, models.ProjectStatusArchived:
		// Only team_lead_id, end_date, budget_spent editable
		restricted.EndDate = updateData.EndDate
		restricted.TeamLeadID = updateData.TeamLeadID
		restricted.BudgetSpent = updateData.BudgetSpent
	}

	return &restricted
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
		log.Printf("❌ ASSIGN_TEAM_LEAD: Invalid project ID: %s", projectIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req struct {
		TeamLeadID uuid.UUID `json:"team_lead_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ ASSIGN_TEAM_LEAD: JSON binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		log.Printf("❌ ASSIGN_TEAM_LEAD: User not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has permission to assign team lead (admin only)
	if !userCtx.HasRole("admin") {
		log.Printf("❌ ASSIGN_TEAM_LEAD: User %s is not admin", userCtx.ID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to assign team lead"})
		return
	}

	log.Printf("📝 ASSIGN_TEAM_LEAD: Assigning team lead %s to project %s by admin %s", req.TeamLeadID, projectID, userCtx.ID)

	if err := h.service.AssignTeamLead(projectID, req.TeamLeadID); err != nil {
		log.Printf("❌ ASSIGN_TEAM_LEAD: Failed to assign team lead: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign team lead", "details": err.Error()})
		return
	}

	log.Printf("✅ ASSIGN_TEAM_LEAD: Successfully assigned team lead %s to project %s", req.TeamLeadID, projectID)
	c.JSON(http.StatusOK, gin.H{"message": "Team lead assigned successfully"})
}

// GetLogistics handles GET /api/projects/:id/logistics
func (h *ProjectHandler) GetLogistics(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
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

	// Check if user is project owner or admin
	isTeamLead, err := h.service.IsTeamLead(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team lead status"})
		return
	}

	if !userCtx.HasRole("admin") && !isTeamLead {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project team lead can view logistics"})
		return
	}

	// Get project details
	project, err := h.service.GetByID(projectID)
	if err != nil || project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Get team members
	teamMembers, err := h.service.GetProjectTeamMembers(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get team members"})
		return
	}

	// Get pending applications
	applicationService := models.NewApplicationService(h.service.GetDB())
	applications, err := applicationService.GetApplicationsByProject(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get applications"})
		return
	}

	// Filter for pending applications
	var pendingApplications []models.Application
	for _, app := range applications {
		if app.Status == "pending" {
			pendingApplications = append(pendingApplications, app)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"project":              project,
		"team_members":         teamMembers,
		"pending_applications": pendingApplications,
	})
}

// UpdateLogisticsRequest represents logistics update request
type UpdateLogisticsRequest struct {
	BudgetTotal *float64 `json:"budget_total"`
	BudgetSpent *float64 `json:"budget_spent"`
}

// UpdateLogistics handles PUT /api/projects/:id/logistics
func (h *ProjectHandler) UpdateLogistics(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req UpdateLogisticsRequest
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

	// Check if user is project owner or admin
	isTeamLead, err := h.service.IsTeamLead(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team lead status"})
		return
	}

	if !userCtx.HasRole("admin") && !isTeamLead {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project team lead can update logistics"})
		return
	}

	// Get existing project
	project, err := h.service.GetByID(projectID)
	if err != nil || project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Update budget fields
	if req.BudgetTotal != nil {
		project.BudgetTotal = req.BudgetTotal
	}
	if req.BudgetSpent != nil {
		project.BudgetSpent = req.BudgetSpent
	}

	// Update project
	if err := h.service.Update(project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update logistics"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// ApproveVolunteerRequest represents volunteer approval request
type ApproveVolunteerRequest struct {
	ApplicationID uuid.UUID `json:"application_id" binding:"required"`
	Approve       bool      `json:"approve"`
}

// ApproveVolunteer handles POST /api/projects/:id/approve-volunteer
func (h *ProjectHandler) ApproveVolunteer(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req ApproveVolunteerRequest
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

	// Check if user is project owner or admin
	isTeamLead, err := h.service.IsTeamLead(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team lead status"})
		return
	}

	if !userCtx.HasRole("admin") && !isTeamLead {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project team lead can approve volunteers"})
		return
	}

	// Get application
	applicationService := models.NewApplicationService(h.service.GetDB())
	application, err := applicationService.GetByID(req.ApplicationID)
	if err != nil || application == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	// Verify application is for this project
	if application.ProjectID != projectID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Application does not belong to this project"})
		return
	}

	// Update application status
	if req.Approve {
		application.Status = "accepted"
		// Note: The database trigger will automatically add to project_team_members
	} else {
		application.Status = "rejected"
	}

	if err := applicationService.Update(application); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update application"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Application processed successfully",
		"application": application,
	})
}

// RemoveVolunteer handles DELETE /api/projects/:id/volunteers/:volunteerId
func (h *ProjectHandler) RemoveVolunteer(c *gin.Context) {
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

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user is project owner or admin
	isTeamLead, err := h.service.IsTeamLead(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team lead status"})
		return
	}

	if !userCtx.HasRole("admin") && !isTeamLead {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project team lead can remove volunteers"})
		return
	}

	// Remove team member
	if err := h.service.RemoveTeamMember(projectID, volunteerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove volunteer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Volunteer removed successfully"})
}

// TransitionProjectStatusRequest represents a project status transition request
type TransitionProjectStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// TransitionProjectStatus handles PUT /api/projects/:id/status
func (h *ProjectHandler) TransitionProjectStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req TransitionProjectStatusRequest
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

	// Validate status
	newStatus := models.ProjectStatus(req.Status)
	validStatuses := []models.ProjectStatus{
		models.ProjectStatusDraft,
		models.ProjectStatusRecruiting,
		models.ProjectStatusActive,
		models.ProjectStatusCompleted,
		models.ProjectStatusArchived,
	}

	validStatus := false
	for _, status := range validStatuses {
		if status == newStatus {
			validStatus = true
			break
		}
	}

	if !validStatus {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project status"})
		return
	}

	log.Printf("🔄 TRANSITION_PROJECT_STATUS: User %s transitioning project %s to %s", userCtx.ID, id, newStatus)

	// Transition project status
	if err := h.service.TransitionProjectStatus(id, newStatus, userCtx.ID); err != nil {
		if err == sql.ErrNoRows {
			if strings.Contains(err.Error(), "permission") {
				c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to transition project status"})
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			}
		} else {
			log.Printf("❌ TRANSITION_PROJECT_STATUS: Failed to transition project: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	log.Printf("✅ TRANSITION_PROJECT_STATUS: Successfully transitioned project %s to %s", id, newStatus)

	// Return updated project
	project, err := h.service.GetByID(id)
	if err != nil {
		log.Printf("⚠️ TRANSITION_PROJECT_STATUS: Failed to fetch updated project: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"message": "Project status updated successfully",
			"status":  newStatus,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project status updated successfully",
		"project": project,
	})
}
