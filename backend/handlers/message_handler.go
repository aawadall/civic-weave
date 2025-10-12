package handlers

import (
	"net/http"
	"strconv"
	"time"

	"civicweave/backend/middleware"
	"civicweave/backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MessageHandler handles message-related requests
type MessageHandler struct {
	messageService *models.MessageService
	projectService *models.ProjectService
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messageService *models.MessageService, projectService *models.ProjectService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
		projectService: projectService,
	}
}

// SendMessageRequest represents message creation request
type SendMessageRequest struct {
	MessageText string `json:"message_text" binding:"required"`
}

// EditMessageRequest represents message edit request
type EditMessageRequest struct {
	MessageText string `json:"message_text" binding:"required"`
}

// ListMessages handles GET /api/projects/:id/messages
func (h *MessageHandler) ListMessages(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get pagination params
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
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

	// Check if user is team member
	isTeamMember, err := h.projectService.IsTeamMember(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team membership"})
		return
	}

	if !isTeamMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team members can view messages"})
		return
	}

	// Get messages
	messages, err := h.messageService.ListByProject(projectID, limit, offset, &userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetRecentMessages handles GET /api/projects/:id/messages/recent
func (h *MessageHandler) GetRecentMessages(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get count param (default 50)
	countStr := c.DefaultQuery("count", "50")
	count, err := strconv.Atoi(countStr)
	if err != nil || count < 1 || count > 100 {
		count = 50
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team members can view messages"})
		return
	}

	// Get recent messages
	messages, err := h.messageService.ListRecentByProject(projectID, count, &userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// GetNewMessages handles GET /api/projects/:id/messages/new
// Returns messages created after a given timestamp (for polling)
func (h *MessageHandler) GetNewMessages(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get 'after' query param (RFC3339 format)
	afterStr := c.Query("after")
	if afterStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'after' timestamp"})
		return
	}

	after, err := time.Parse(time.RFC3339, afterStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timestamp format (use RFC3339)"})
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team members can view messages"})
		return
	}

	// Get new messages
	messages, err := h.messageService.GetMessagesAfter(projectID, after, &userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// SendMessage handles POST /api/projects/:id/messages
func (h *MessageHandler) SendMessage(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req SendMessageRequest
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

	// Check if user is team member
	isTeamMember, err := h.projectService.IsTeamMember(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team membership"})
		return
	}

	if !isTeamMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team members can send messages"})
		return
	}

	// Create message
	message := &models.ProjectMessage{
		ProjectID:   projectID,
		SenderID:    userCtx.ID,
		MessageText: req.MessageText,
	}

	if err := h.messageService.Create(message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	// TODO: Publish to Redis Pub/Sub for real-time updates
	// This will be added in the messaging_service.go

	c.JSON(http.StatusCreated, message)
}

// EditMessage handles PUT /api/messages/:id
func (h *MessageHandler) EditMessage(c *gin.Context) {
	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	var req EditMessageRequest
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

	// Get message
	message, err := h.messageService.GetByID(messageID)
	if err != nil || message == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Check if message is soft-deleted
	if message.DeletedAt != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Check if user can edit (sender and within 15 minutes)
	canEdit, err := h.messageService.CanUserEdit(messageID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check edit permissions"})
		return
	}

	if !canEdit {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only edit your own messages within 15 minutes"})
		return
	}

	// Update message
	message.MessageText = req.MessageText
	if err := h.messageService.Update(message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update message"})
		return
	}

	c.JSON(http.StatusOK, message)
}

// DeleteMessage handles DELETE /api/messages/:id
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get message
	message, err := h.messageService.GetByID(messageID)
	if err != nil || message == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Check if message is already deleted
	if message.DeletedAt != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Check permissions (sender, project owner, or admin)
	isSender := message.SenderID == userCtx.ID
	isProjectOwner, err := h.projectService.IsTeamLead(message.ProjectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check project ownership"})
		return
	}

	isAdmin := userCtx.HasRole("admin")

	if !isSender && !isProjectOwner && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the sender, project lead, or admin can delete messages"})
		return
	}

	// Soft delete message
	if err := h.messageService.SoftDelete(messageID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message deleted successfully"})
}

// MarkMessageAsRead handles POST /api/messages/:id/read
func (h *MessageHandler) MarkMessageAsRead(c *gin.Context) {
	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Mark as read
	if err := h.messageService.MarkAsRead(messageID, userCtx.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark message as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message marked as read"})
}

// MarkAllAsRead handles POST /api/projects/:id/messages/read-all
func (h *MessageHandler) MarkAllAsRead(c *gin.Context) {
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team members can mark messages as read"})
		return
	}

	// Mark all as read
	if err := h.messageService.MarkAllAsRead(projectID, userCtx.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark messages as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All messages marked as read"})
}

// GetUnreadCount handles GET /api/projects/:id/messages/unread-count
func (h *MessageHandler) GetUnreadCount(c *gin.Context) {
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

	// Get unread count
	count, err := h.messageService.GetUnreadCount(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// GetAllUnreadCounts handles GET /api/messages/unread-counts
func (h *MessageHandler) GetAllUnreadCounts(c *gin.Context) {
	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get unread counts for all projects
	counts, err := h.messageService.GetUnreadCountsByUser(userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unread counts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_counts": counts})
}

