package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"civicweave/backend/middleware"
	"civicweave/backend/models"

	"github.com/gin-gonic/gin"
)

// UserDashboardHandler handles user dashboard data aggregation
type UserDashboardHandler struct {
	projectService   *models.ProjectService
	taskService      *models.TaskService
	messageService   *models.MessageService
	broadcastService *models.BroadcastService
	resourceService  *models.ResourceService
}

// NewUserDashboardHandler creates a new user dashboard handler
func NewUserDashboardHandler(
	projectService *models.ProjectService,
	taskService *models.TaskService,
	messageService *models.MessageService,
	broadcastService *models.BroadcastService,
	resourceService *models.ResourceService,
) *UserDashboardHandler {
	return &UserDashboardHandler{
		projectService:   projectService,
		taskService:      taskService,
		messageService:   messageService,
		broadcastService: broadcastService,
		resourceService:  resourceService,
	}
}

// DashboardData represents aggregated dashboard data
type DashboardData struct {
	Projects   []models.ProjectWithDetails   `json:"projects"`
	Tasks      []models.ProjectTask          `json:"tasks"`
	Messages   models.UniversalUnreadCount   `json:"messages"`
	Broadcasts []models.BroadcastWithAuthor  `json:"broadcasts"`
	Resources  []models.ResourceWithUploader `json:"resources"`
	Stats      DashboardStats                `json:"stats"`
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalProjects    int `json:"total_projects"`
	ActiveTasks      int `json:"active_tasks"`
	OverdueTasks     int `json:"overdue_tasks"`
	UnreadMessages   int `json:"unread_messages"`
	UnreadBroadcasts int `json:"unread_broadcasts"`
	RecentResources  int `json:"recent_resources"`
}

// GetUserProjects handles GET /api/users/me/projects
func (h *UserDashboardHandler) GetUserProjects(c *gin.Context) {
	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get user's enrolled projects
	projects, err := h.projectService.GetUserEnrolledProjects(userCtx.ID)
	if err != nil {
		log.Printf("❌ GET_USER_PROJECTS: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user projects"})
		return
	}

	log.Printf("✅ GET_USER_PROJECTS: Successfully fetched %d projects for user %s", len(projects), userCtx.ID)

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"count":    len(projects),
	})
}

// GetUserTasks handles GET /api/users/me/tasks
func (h *UserDashboardHandler) GetUserTasks(c *gin.Context) {
	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get user's volunteer profile
	volunteerService := models.NewVolunteerService(h.projectService.GetDB())
	volunteer, err := volunteerService.GetByUserID(userCtx.ID)
	if err != nil || volunteer == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Volunteer profile not found"})
		return
	}

	// Get user's assigned tasks
	tasks, err := h.taskService.ListByAssignee(volunteer.ID)
	if err != nil {
		log.Printf("❌ GET_USER_TASKS: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user tasks"})
		return
	}

	log.Printf("✅ GET_USER_TASKS: Successfully fetched %d tasks for user %s", len(tasks), userCtx.ID)

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"count": len(tasks),
	})
}

// GetDashboardData handles GET /api/users/me/dashboard
func (h *UserDashboardHandler) GetDashboardData(c *gin.Context) {
	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get query parameters for limits
	projectsLimitStr := c.DefaultQuery("projects_limit", "5")
	tasksLimitStr := c.DefaultQuery("tasks_limit", "10")
	broadcastsLimitStr := c.DefaultQuery("broadcasts_limit", "5")
	resourcesLimitStr := c.DefaultQuery("resources_limit", "5")

	projectsLimit, _ := strconv.Atoi(projectsLimitStr)
	if projectsLimit < 1 || projectsLimit > 20 {
		projectsLimit = 5
	}

	tasksLimit, _ := strconv.Atoi(tasksLimitStr)
	if tasksLimit < 1 || tasksLimit > 50 {
		tasksLimit = 10
	}

	broadcastsLimit, _ := strconv.Atoi(broadcastsLimitStr)
	if broadcastsLimit < 1 || broadcastsLimit > 20 {
		broadcastsLimit = 5
	}

	resourcesLimit, _ := strconv.Atoi(resourcesLimitStr)
	if resourcesLimit < 1 || resourcesLimit > 20 {
		resourcesLimit = 5
	}

	// Get user's volunteer profile
	volunteerService := models.NewVolunteerService(h.projectService.GetDB())
	volunteer, err := volunteerService.GetByUserID(userCtx.ID)
	if err != nil || volunteer == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Volunteer profile not found"})
		return
	}

	// Get user roles for broadcast filtering
	userService := models.NewUserService(h.projectService.GetDB())
	roles, err := userService.GetUserRoles(userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	var roleNames []string
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	// Fetch all dashboard data in parallel
	type projectsResult struct {
		projects []models.ProjectWithDetails
		err      error
	}
	type tasksResult struct {
		tasks []models.ProjectTask
		err   error
	}
	type messagesResult struct {
		messages *models.UniversalUnreadCount
		err      error
	}
	type broadcastsResult struct {
		broadcasts []models.BroadcastWithAuthor
		err        error
	}
	type resourcesResult struct {
		resources []models.ResourceWithUploader
		err       error
	}

	projectsChan := make(chan projectsResult, 1)
	tasksChan := make(chan tasksResult, 1)
	messagesChan := make(chan messagesResult, 1)
	broadcastsChan := make(chan broadcastsResult, 1)
	resourcesChan := make(chan resourcesResult, 1)

	// Fetch projects
	go func() {
		projects, err := h.projectService.GetUserEnrolledProjects(userCtx.ID)
		if len(projects) > projectsLimit {
			projects = projects[:projectsLimit]
		}
		projectsChan <- projectsResult{projects: projects, err: err}
	}()

	// Fetch tasks
	go func() {
		tasks, err := h.taskService.ListByAssignee(volunteer.ID)
		if len(tasks) > tasksLimit {
			tasks = tasks[:tasksLimit]
		}
		tasksChan <- tasksResult{tasks: tasks, err: err}
	}()

	// Fetch messages
	go func() {
		messages, err := h.messageService.GetUniversalUnreadCount(userCtx.ID)
		messagesChan <- messagesResult{messages: messages, err: err}
	}()

	// Fetch broadcasts
	go func() {
		userRole := "volunteer" // default
		if len(roleNames) > 0 {
			userRole = roleNames[0]
		}
		broadcasts, err := h.broadcastService.List(userCtx.ID, userRole, broadcastsLimit, 0)
		broadcastsChan <- broadcastsResult{broadcasts: broadcasts, err: err}
	}()

	// Fetch resources
	go func() {
		filters := models.ResourceFilters{}
		resources, err := h.resourceService.List(filters, resourcesLimit, 0)
		resourcesChan <- resourcesResult{resources: resources, err: err}
	}()

	// Collect results
	projectsRes := <-projectsChan
	tasksRes := <-tasksChan
	messagesRes := <-messagesChan
	broadcastsRes := <-broadcastsChan
	resourcesRes := <-resourcesChan

	// Check for errors
	if projectsRes.err != nil {
		log.Printf("❌ GET_DASHBOARD_DATA: Projects error: %v", projectsRes.err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get projects"})
		return
	}
	if tasksRes.err != nil {
		log.Printf("❌ GET_DASHBOARD_DATA: Tasks error: %v", tasksRes.err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}
	if messagesRes.err != nil {
		log.Printf("❌ GET_DASHBOARD_DATA: Messages error: %v", messagesRes.err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}
	if broadcastsRes.err != nil {
		// Check if the error is due to missing table
		if strings.Contains(broadcastsRes.err.Error(), "does not exist") {
			log.Printf("⚠️ GET_DASHBOARD_DATA: Broadcast table not found, using empty list")
			broadcastsRes.broadcasts = []models.BroadcastWithAuthor{}
		} else {
			log.Printf("❌ GET_DASHBOARD_DATA: Broadcasts error: %v", broadcastsRes.err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get broadcasts"})
			return
		}
	}
	if resourcesRes.err != nil {
		// Check if the error is due to missing table
		if strings.Contains(resourcesRes.err.Error(), "does not exist") {
			log.Printf("⚠️ GET_DASHBOARD_DATA: Resources table not found, using empty list")
			resourcesRes.resources = []models.ResourceWithUploader{}
		} else {
			log.Printf("❌ GET_DASHBOARD_DATA: Resources error: %v", resourcesRes.err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get resources"})
			return
		}
	}

	// Calculate stats
	stats := DashboardStats{
		TotalProjects:    len(projectsRes.projects),
		ActiveTasks:      len(tasksRes.tasks),
		UnreadMessages:   messagesRes.messages.Total,
		UnreadBroadcasts: len(broadcastsRes.broadcasts),
		RecentResources:  len(resourcesRes.resources),
	}

	// Count overdue tasks
	overdueCount := 0
	for _, task := range tasksRes.tasks {
		if task.DueDate != nil && task.DueDate.Before(time.Now()) && task.Status != "done" {
			overdueCount++
		}
	}
	stats.OverdueTasks = overdueCount

	// Build dashboard data
	dashboardData := DashboardData{
		Projects:   projectsRes.projects,
		Tasks:      tasksRes.tasks,
		Messages:   *messagesRes.messages,
		Broadcasts: broadcastsRes.broadcasts,
		Resources:  resourcesRes.resources,
		Stats:      stats,
	}

	log.Printf("✅ GET_DASHBOARD_DATA: Successfully fetched dashboard data for user %s", userCtx.ID)

	c.JSON(http.StatusOK, dashboardData)
}
