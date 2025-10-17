package handlers

import (
	"net/http"
	"time"

	"civicweave/backend/middleware"
	"civicweave/backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TaskHandler handles task-related requests
type TaskHandler struct {
	taskService      *models.TaskService
	projectService   *models.ProjectService
	volunteerService *models.VolunteerService
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(taskService *models.TaskService, projectService *models.ProjectService, volunteerService *models.VolunteerService) *TaskHandler {
	return &TaskHandler{
		taskService:      taskService,
		projectService:   projectService,
		volunteerService: volunteerService,
	}
}

// CreateTaskRequest represents task creation request
type CreateTaskRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	AssigneeID  *uuid.UUID `json:"assignee_id"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"due_date"`
	Labels      []string   `json:"labels"`
}

// UpdateTaskRequest represents task update request
type UpdateTaskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	AssigneeID  *uuid.UUID `json:"assignee_id"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"due_date"`
	Labels      []string   `json:"labels"`
}

// AddUpdateRequest represents adding a progress update
type AddUpdateRequest struct {
	UpdateText string `json:"update_text" binding:"required"`
}

// ListTasks handles GET /api/projects/:id/tasks
func (h *TaskHandler) ListTasks(c *gin.Context) {
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

	// Check if user is team member
	isTeamMember, err := h.projectService.IsTeamMember(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team membership"})
		return
	}

	if !isTeamMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team members can view tasks"})
		return
	}

	// Check if user is project owner
	isProjectOwner, err := h.projectService.IsTeamLead(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check project ownership"})
		return
	}

	// Admin can see all tasks too
	if userCtx.HasRole("admin") {
		isProjectOwner = true
	}

	// Get tasks (filtered based on permissions)
	tasks, err := h.taskService.ListByProject(projectID, &userCtx.ID, isProjectOwner)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks":            tasks,
		"is_project_owner": isProjectOwner,
	})
}

// ListUnassignedTasks handles GET /api/projects/:id/tasks/unassigned
func (h *TaskHandler) ListUnassignedTasks(c *gin.Context) {
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

	// Check if user is team member
	isTeamMember, err := h.projectService.IsTeamMember(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team membership"})
		return
	}

	if !isTeamMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team members can view tasks"})
		return
	}

	// Get unassigned tasks
	tasks, err := h.taskService.ListUnassignedByProject(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unassigned tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// CreateTask handles POST /api/projects/:id/tasks
func (h *TaskHandler) CreateTask(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req CreateTaskRequest
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
	isTeamLead, err := h.projectService.IsTeamLead(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team lead status"})
		return
	}

	if !userCtx.HasRole("admin") && !isTeamLead {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project team lead can create tasks"})
		return
	}

	// Set default priority if not specified
	priority := models.TaskPriorityMedium
	if req.Priority != "" {
		priority = models.TaskPriority(req.Priority)
	}

	// Create task
	task := &models.ProjectTask{
		ProjectID:   projectID,
		Title:       req.Title,
		Description: req.Description,
		AssigneeID:  req.AssigneeID,
		CreatedByID: userCtx.ID,
		Status:      models.TaskStatusTodo,
		Priority:    priority,
		DueDate:     req.DueDate,
		Labels:      req.Labels,
	}

	if err := h.taskService.Create(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// GetTask handles GET /api/tasks/:id
func (h *TaskHandler) GetTask(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get task with updates
	taskWithUpdates, err := h.taskService.GetTaskWithUpdates(taskID)
	if err != nil || taskWithUpdates == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Check if user has access (team member, project owner, or admin)
	isTeamMember, err := h.projectService.IsTeamMember(taskWithUpdates.ProjectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team membership"})
		return
	}

	if !isTeamMember && !userCtx.HasRole("admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, taskWithUpdates)
}

// UpdateTask handles PUT /api/tasks/:id
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req UpdateTaskRequest
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

	// Get existing task
	task, err := h.taskService.GetByID(taskID)
	if err != nil || task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Check permissions
	isTeamLead, err := h.projectService.IsTeamLead(task.ProjectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team lead status"})
		return
	}

	isAssignee := task.AssigneeID != nil && *task.AssigneeID == userCtx.ID
	isAdmin := userCtx.HasRole("admin")

	// Team lead and admin can update everything
	// Assignee can only update status
	if !isAdmin && !isTeamLead {
		if !isAssignee {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only project team lead or assignee can update task"})
			return
		}
		// Assignee can only update status
		if req.Status != "" {
			if err := h.taskService.UpdateStatus(taskID, models.TaskStatus(req.Status)); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task status"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Task status updated successfully"})
			return
		}
	}

	// Update fields (for team lead and admin)
	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.AssigneeID != nil {
		task.AssigneeID = req.AssigneeID
	}
	if req.Status != "" {
		task.Status = models.TaskStatus(req.Status)
	}
	if req.Priority != "" {
		task.Priority = models.TaskPriority(req.Priority)
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}
	if req.Labels != nil {
		task.Labels = req.Labels
	}

	if err := h.taskService.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// SelfAssignTask handles POST /api/tasks/:id/assign
func (h *TaskHandler) SelfAssignTask(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get task
	task, err := h.taskService.GetByID(taskID)
	if err != nil || task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Check if user is team member
	isTeamMember, err := h.projectService.IsTeamMember(task.ProjectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team membership"})
		return
	}

	if !isTeamMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team members can self-assign tasks"})
		return
	}

	// Check if task is already assigned
	if task.AssigneeID != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task is already assigned"})
		return
	}

	// Get volunteer ID for user
	volunteer, err := h.volunteerService.GetByUserID(userCtx.ID)
	if err != nil || volunteer == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Volunteer profile not found"})
		return
	}

	// Assign task
	if err := h.taskService.AssignToVolunteer(taskID, volunteer.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task assigned successfully"})
}

// AddTaskUpdate handles POST /api/tasks/:id/updates
func (h *TaskHandler) AddTaskUpdate(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req AddUpdateRequest
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

	// Get task
	task, err := h.taskService.GetByID(taskID)
	if err != nil || task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Get volunteer ID for user
	volunteer, err := h.volunteerService.GetByUserID(userCtx.ID)
	if err != nil || volunteer == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Volunteer profile not found"})
		return
	}

	// Check if user is the assignee
	if task.AssigneeID == nil || *task.AssigneeID != volunteer.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the assignee can add updates"})
		return
	}

	// Add update
	update := &models.TaskUpdate{
		TaskID:      taskID,
		VolunteerID: volunteer.ID,
		UpdateText:  req.UpdateText,
	}

	if err := h.taskService.AddUpdate(update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add update"})
		return
	}

	c.JSON(http.StatusCreated, update)
}

// DeleteTask handles DELETE /api/tasks/:id
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get task
	task, err := h.taskService.GetByID(taskID)
	if err != nil || task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Check if user is project owner or admin
	isTeamLead, err := h.projectService.IsTeamLead(task.ProjectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team lead status"})
		return
	}

	if !userCtx.HasRole("admin") && !isTeamLead {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project team lead can delete tasks"})
		return
	}

	// Delete task
	if err := h.taskService.Delete(taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// AddTaskComment handles POST /api/tasks/:id/comments
func (h *TaskHandler) AddTaskComment(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req struct {
		CommentText string `json:"comment_text" binding:"required"`
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

	// Get task to verify it exists
	task, err := h.taskService.GetByID(taskID)
	if err != nil || task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Check if user is team member
	isTeamMember, err := h.projectService.IsTeamMember(task.ProjectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team membership"})
		return
	}

	if !isTeamMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team members can comment on tasks"})
		return
	}

	// Create comment
	comment := &models.TaskComment{
		TaskID:      taskID,
		UserID:      userCtx.ID,
		CommentText: req.CommentText,
	}

	commentService := models.NewTaskCommentService(h.taskService.(*models.TaskService).db)
	if err := commentService.Create(comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add comment"})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// GetTaskComments handles GET /api/tasks/:id/comments
func (h *TaskHandler) GetTaskComments(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get task to verify it exists and check permissions
	task, err := h.taskService.GetByID(taskID)
	if err != nil || task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Check if user is team member
	isTeamMember, err := h.projectService.IsTeamMember(task.ProjectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team membership"})
		return
	}

	if !isTeamMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team members can view task comments"})
		return
	}

	// Get comments
	commentService := models.NewTaskCommentService(h.taskService.(*models.TaskService).db)
	comments, err := commentService.GetByTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get comments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})
}

// LogTaskTime handles POST /api/tasks/:id/time-logs
func (h *TaskHandler) LogTaskTime(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req struct {
		Hours       float64   `json:"hours" binding:"required,min=0.1"`
		LogDate     time.Time `json:"log_date" binding:"required"`
		Description string    `json:"description"`
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

	// Get task to verify it exists
	task, err := h.taskService.GetByID(taskID)
	if err != nil || task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Check if user is team member
	isTeamMember, err := h.projectService.IsTeamMember(task.ProjectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team membership"})
		return
	}

	if !isTeamMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team members can log time"})
		return
	}

	// Get volunteer ID for user
	volunteer, err := h.volunteerService.GetByUserID(userCtx.ID)
	if err != nil || volunteer == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Volunteer profile not found"})
		return
	}

	// Create time log
	timeLog := &models.TaskTimeLog{
		TaskID:      taskID,
		VolunteerID: volunteer.ID,
		Hours:       req.Hours,
		LogDate:     req.LogDate,
		Description: req.Description,
	}

	timeLogService := models.NewTaskTimeLogService(h.taskService.(*models.TaskService).db)
	if err := timeLogService.Create(timeLog); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log time"})
		return
	}

	c.JSON(http.StatusCreated, timeLog)
}

// GetTaskTimeLogs handles GET /api/tasks/:id/time-logs
func (h *TaskHandler) GetTaskTimeLogs(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get task to verify it exists and check permissions
	task, err := h.taskService.GetByID(taskID)
	if err != nil || task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Check if user is team member
	isTeamMember, err := h.projectService.IsTeamMember(task.ProjectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team membership"})
		return
	}

	if !isTeamMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team members can view time logs"})
		return
	}

	// Get time logs
	timeLogService := models.NewTaskTimeLogService(h.taskService.(*models.TaskService).db)
	timeLogs, err := timeLogService.GetByTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get time logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"time_logs": timeLogs})
}

// MarkTaskBlocked handles POST /api/tasks/:id/mark-blocked
func (h *TaskHandler) MarkTaskBlocked(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
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

	// Get task to verify it exists
	task, err := h.taskService.GetByID(taskID)
	if err != nil || task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Check if user is the assignee
	if task.AssigneeID == nil || *task.AssigneeID != userCtx.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the task assignee can mark it as blocked"})
		return
	}

	// Mark as blocked
	if err := h.taskService.MarkAsBlocked(taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark task as blocked"})
		return
	}

	// Create notification message
	messageService := models.NewMessageService(h.taskService.(*models.TaskService).db)
	messageText := "🚫 Task blocked"
	if req.Reason != "" {
		messageText += ": " + req.Reason
	}
	if err := messageService.CreateTaskNotification(task.ProjectID, userCtx.ID, taskID, "task_blocked", messageText); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task marked as blocked"})
}

// RequestTaskTakeover handles POST /api/tasks/:id/request-takeover
func (h *TaskHandler) RequestTaskTakeover(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
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

	// Get task to verify it exists
	task, err := h.taskService.GetByID(taskID)
	if err != nil || task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Check if user is the assignee
	if task.AssigneeID == nil || *task.AssigneeID != userCtx.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the task assignee can request takeover"})
		return
	}

	// Request takeover
	if err := h.taskService.RequestTakeover(taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to request task takeover"})
		return
	}

	// Create notification message
	messageService := models.NewMessageService(h.taskService.(*models.TaskService).db)
	messageText := "🔄 Task takeover requested"
	if req.Reason != "" {
		messageText += ": " + req.Reason
	}
	if err := messageService.CreateTaskNotification(task.ProjectID, userCtx.ID, taskID, "task_takeover", messageText); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task takeover requested"})
}

// MarkTaskDone handles POST /api/tasks/:id/mark-done
func (h *TaskHandler) MarkTaskDone(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req struct {
		CompletionNote string `json:"completion_note"`
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

	// Get task to verify it exists
	task, err := h.taskService.GetByID(taskID)
	if err != nil || task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Check if user is the assignee
	if task.AssigneeID == nil || *task.AssigneeID != userCtx.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the task assignee can mark it as done"})
		return
	}

	// Mark as done
	if err := h.taskService.MarkAsDone(taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark task as done"})
		return
	}

	// Create notification message
	messageService := models.NewMessageService(h.taskService.(*models.TaskService).db)
	messageText := "✅ Task completed"
	if req.CompletionNote != "" {
		messageText += ": " + req.CompletionNote
	}
	if err := messageService.CreateTaskNotification(task.ProjectID, userCtx.ID, taskID, "task_done", messageText); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task marked as done"})
}

