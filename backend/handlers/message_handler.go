package handlers

import (
	"log"
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
	userService    *models.UserService
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messageService *models.MessageService, projectService *models.ProjectService, userService *models.UserService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
		projectService: projectService,
		userService:    userService,
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

	// Check if user is team member or team lead
	isTeamMember, err := h.projectService.IsTeamMember(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team membership"})
		return
	}

	isTeamLead, err := h.projectService.IsTeamLead(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team lead status"})
		return
	}

	if !isTeamMember && !isTeamLead {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team members and team leads can view messages"})
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

	// Check if user is team member or team lead
	isTeamMember, err := h.projectService.IsTeamMember(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team membership"})
		return
	}

	isTeamLead, err := h.projectService.IsTeamLead(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team lead status"})
		return
	}

	if !isTeamMember && !isTeamLead {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team members and team leads can view messages"})
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

	// Check if user is team member or team lead
	isTeamMember, err := h.projectService.IsTeamMember(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team membership"})
		return
	}

	isTeamLead, err := h.projectService.IsTeamLead(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team lead status"})
		return
	}

	if !isTeamMember && !isTeamLead {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team members and team leads can view messages"})
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

	// Check if user is team member or team lead
	isTeamMember, err := h.projectService.IsTeamMember(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team membership"})
		return
	}

	isTeamLead, err := h.projectService.IsTeamLead(projectID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team lead status"})
		return
	}

	if !isTeamMember && !isTeamLead {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team members and team leads can send messages"})
		return
	}

	// Create message
	message := &models.ProjectMessage{
		ProjectID:   &projectID,
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
	var isProjectOwner bool
	if message.ProjectID != nil {
		isProjectOwner, err = h.projectService.IsTeamLead(*message.ProjectID, userCtx.ID)
	}
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

// SendUniversalMessageRequest represents universal message creation request
type SendUniversalMessageRequest struct {
	RecipientType string  `json:"recipient_type" binding:"required"` // "user", "team", "project"
	RecipientID   string  `json:"recipient_id" binding:"required"`
	Subject       *string `json:"subject,omitempty"`
	MessageText   string  `json:"message_text" binding:"required"`
}

// SendUniversalMessage handles POST /api/messages
func (h *MessageHandler) SendUniversalMessage(c *gin.Context) {
	var req SendUniversalMessageRequest
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

	// Parse recipient ID
	recipientID, err := uuid.Parse(req.RecipientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipient ID"})
		return
	}

	// Create message based on recipient type
	message := &models.ProjectMessage{
		SenderID:    userCtx.ID,
		Subject:     req.Subject,
		MessageText: req.MessageText,
		MessageType: "general",
	}

	log.Printf("DEBUG: Creating message with recipient type: %s, recipient ID: %s", req.RecipientType, req.RecipientID)

	switch req.RecipientType {
	case "user":
		message.RecipientUserID = &recipientID
		message.MessageScope = "user_to_user"
	case "team":
		message.RecipientTeamID = &recipientID
		message.MessageScope = "user_to_team"
	case "project":
		message.ProjectID = &recipientID
		message.MessageScope = "project"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipient type"})
		return
	}

	// Validate recipient exists and user has permission
	if req.RecipientType == "user" {
		// Check if recipient user exists
		recipientUser, err := h.userService.GetByID(recipientID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate recipient"})
			return
		}
		if recipientUser == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Recipient user does not exist"})
			return
		}
		// Prevent self-messaging
		if recipientID == userCtx.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot send message to yourself"})
			return
		}
	} else if req.RecipientType == "team" || req.RecipientType == "project" {
		// Check if user is team member
		isTeamMember, err := h.projectService.IsTeamMember(recipientID, userCtx.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team membership"})
			return
		}

		if !isTeamMember {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only team members can send messages to this project/team"})
			return
		}
	}

	log.Printf("DEBUG: About to call CreateUniversalMessage")
	if err := h.messageService.CreateUniversalMessage(message); err != nil {
		log.Printf("DEBUG: CreateUniversalMessage error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}
	log.Printf("DEBUG: CreateUniversalMessage succeeded")

	c.JSON(http.StatusCreated, message)
}

// GetInbox handles GET /api/messages/inbox
func (h *MessageHandler) GetInbox(c *gin.Context) {
	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
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

	// Get inbox messages
	messages, err := h.messageService.GetInbox(userCtx.ID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get inbox messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"limit":    limit,
		"offset":   offset,
		"count":    len(messages),
	})
}

// GetSentMessages handles GET /api/messages/sent
func (h *MessageHandler) GetSentMessages(c *gin.Context) {
	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
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

	// Get sent messages
	messages, err := h.messageService.GetSentMessages(userCtx.ID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sent messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"limit":    limit,
		"offset":   offset,
		"count":    len(messages),
	})
}

// GetConversations handles GET /api/messages/conversations
func (h *MessageHandler) GetConversations(c *gin.Context) {
	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get pagination params
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

	// Get conversations
	conversations, err := h.messageService.GetConversations(userCtx.ID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get conversations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversations": conversations,
		"limit":         limit,
		"offset":        offset,
		"count":         len(conversations),
	})
}

// GetConversation handles GET /api/messages/conversations/:id
func (h *MessageHandler) GetConversation(c *gin.Context) {
	conversationIDStr := c.Param("id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
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

	// Get conversation messages
	messages, err := h.messageService.GetConversation(conversationID, userCtx.ID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get conversation messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"limit":    limit,
		"offset":   offset,
		"count":    len(messages),
	})
}

// GetUniversalUnreadCount handles GET /api/messages/unread-count
func (h *MessageHandler) GetUniversalUnreadCount(c *gin.Context) {
	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get universal unread count
	count, err := h.messageService.GetUniversalUnreadCount(userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, count)
}

// SearchRecipients handles GET /api/messages/recipients/search
func (h *MessageHandler) SearchRecipients(c *gin.Context) {
	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get search query
	query := c.Query("q")
	if len(query) < 2 {
		c.JSON(http.StatusOK, gin.H{
			"users":    []interface{}{},
			"projects": []interface{}{},
		})
		return
	}

	// Search users and projects
	users, err := h.messageService.SearchUsers(query, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search users"})
		return
	}

	projects, err := h.messageService.SearchUserProjects(userCtx.ID, query, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search projects"})
		return
	}

	// Ensure we always return arrays, never null
	if users == nil {
		users = []models.SearchUser{}
	}
	if projects == nil {
		projects = []models.SearchProject{}
	}

	c.JSON(http.StatusOK, gin.H{
		"users":    users,
		"projects": projects,
	})
}
