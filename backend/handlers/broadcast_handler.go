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
	"github.com/google/uuid"
)

// BroadcastHandler handles broadcast-related requests
type BroadcastHandler struct {
	service *models.BroadcastService
}

// NewBroadcastHandler creates a new broadcast handler
func NewBroadcastHandler(service *models.BroadcastService) *BroadcastHandler {
	return &BroadcastHandler{
		service: service,
	}
}

// CreateBroadcastRequest represents broadcast creation request
type CreateBroadcastRequest struct {
	Title          string     `json:"title" binding:"required"`
	Content        string     `json:"content" binding:"required"`
	TargetAudience string     `json:"target_audience" binding:"required"`
	Priority       string     `json:"priority"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

// UpdateBroadcastRequest represents broadcast update request
type UpdateBroadcastRequest struct {
	Title          string     `json:"title"`
	Content        string     `json:"content"`
	TargetAudience string     `json:"target_audience"`
	Priority       string     `json:"priority"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

// ListBroadcasts handles GET /api/broadcasts
func (h *BroadcastHandler) ListBroadcasts(c *gin.Context) {
	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
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

	// Get user primary role (use the first role or default to volunteer)
	userService := models.NewUserService(h.service.GetDB())
	roles, err := userService.GetUserRoles(userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	userRole := "volunteer" // default
	if len(roles) > 0 {
		userRole = roles[0].Name
	}

	// Get broadcasts
	broadcasts, err := h.service.List(userCtx.ID, userRole, limit, offset)
	if err != nil {
		// Check if the error is due to missing table
		if strings.Contains(err.Error(), "does not exist") {
			log.Printf("⚠️ LIST_BROADCASTS: Broadcast table not found, returning empty list")
			c.JSON(http.StatusOK, gin.H{
				"broadcasts": []interface{}{},
				"count":      0,
			})
			return
		}
		log.Printf("❌ LIST_BROADCASTS: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get broadcasts"})
		return
	}

	log.Printf("✅ LIST_BROADCASTS: Successfully fetched %d broadcasts", len(broadcasts))

	c.JSON(http.StatusOK, gin.H{
		"broadcasts": broadcasts,
		"limit":      limit,
		"offset":     offset,
		"count":      len(broadcasts),
	})
}

// GetBroadcast handles GET /api/broadcasts/:id
func (h *BroadcastHandler) GetBroadcast(c *gin.Context) {
	broadcastIDStr := c.Param("id")
	broadcastID, err := uuid.Parse(broadcastIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid broadcast ID"})
		return
	}

	// Get broadcast
	broadcast, err := h.service.GetByID(broadcastID)
	if err != nil {
		log.Printf("❌ GET_BROADCAST: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get broadcast"})
		return
	}

	if broadcast == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Broadcast not found"})
		return
	}

	log.Printf("✅ GET_BROADCAST: Successfully fetched broadcast %s", broadcastID)

	c.JSON(http.StatusOK, broadcast)
}

// CreateBroadcast handles POST /api/broadcasts
func (h *BroadcastHandler) CreateBroadcast(c *gin.Context) {
	var req CreateBroadcastRequest
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

	// Check if user is admin
	userService := models.NewUserService(h.service.GetDB())
	roles, err := userService.GetUserRoles(userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	isAdmin := false
	for _, role := range roles {
		if role.Name == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can create broadcasts"})
		return
	}

	// Set default priority if not specified
	priority := "normal"
	if req.Priority != "" {
		priority = req.Priority
	}

	// Create broadcast
	broadcast := &models.BroadcastMessage{
		Title:          req.Title,
		Content:        req.Content,
		AuthorID:       userCtx.ID,
		TargetAudience: req.TargetAudience,
		Priority:       priority,
		ExpiresAt:      req.ExpiresAt,
	}

	if err := h.service.Create(broadcast); err != nil {
		log.Printf("❌ CREATE_BROADCAST: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create broadcast"})
		return
	}

	log.Printf("✅ CREATE_BROADCAST: Successfully created broadcast %s", broadcast.ID)

	c.JSON(http.StatusCreated, broadcast)
}

// UpdateBroadcast handles PUT /api/broadcasts/:id
func (h *BroadcastHandler) UpdateBroadcast(c *gin.Context) {
	broadcastIDStr := c.Param("id")
	broadcastID, err := uuid.Parse(broadcastIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid broadcast ID"})
		return
	}

	var req UpdateBroadcastRequest
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

	// Check if user is author or admin
	isAuthor, err := h.service.IsAuthor(broadcastID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check authorship"})
		return
	}

	// Check if user is admin
	userService := models.NewUserService(h.service.GetDB())
	roles, err := userService.GetUserRoles(userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	isAdmin := false
	for _, role := range roles {
		if role.Name == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAuthor && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only broadcast author or admin can update"})
		return
	}

	// Get existing broadcast
	broadcast, err := h.service.GetByID(broadcastID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get broadcast"})
		return
	}

	if broadcast == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Broadcast not found"})
		return
	}

	// Update fields
	if req.Title != "" {
		broadcast.Title = req.Title
	}
	if req.Content != "" {
		broadcast.Content = req.Content
	}
	if req.TargetAudience != "" {
		broadcast.TargetAudience = req.TargetAudience
	}
	if req.Priority != "" {
		broadcast.Priority = req.Priority
	}
	if req.ExpiresAt != nil {
		broadcast.ExpiresAt = req.ExpiresAt
	}

	if err := h.service.Update(broadcast); err != nil {
		log.Printf("❌ UPDATE_BROADCAST: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update broadcast"})
		return
	}

	log.Printf("✅ UPDATE_BROADCAST: Successfully updated broadcast %s", broadcastID)

	c.JSON(http.StatusOK, broadcast)
}

// DeleteBroadcast handles DELETE /api/broadcasts/:id
func (h *BroadcastHandler) DeleteBroadcast(c *gin.Context) {
	broadcastIDStr := c.Param("id")
	broadcastID, err := uuid.Parse(broadcastIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid broadcast ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user is author or admin
	isAuthor, err := h.service.IsAuthor(broadcastID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check authorship"})
		return
	}

	// Check if user is admin
	userService := models.NewUserService(h.service.GetDB())
	roles, err := userService.GetUserRoles(userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	isAdmin := false
	for _, role := range roles {
		if role.Name == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAuthor && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only broadcast author or admin can delete"})
		return
	}

	if err := h.service.SoftDelete(broadcastID); err != nil {
		log.Printf("❌ DELETE_BROADCAST: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete broadcast"})
		return
	}

	log.Printf("✅ DELETE_BROADCAST: Successfully deleted broadcast %s", broadcastID)

	c.JSON(http.StatusOK, gin.H{"message": "Broadcast deleted successfully"})
}

// MarkBroadcastAsRead handles POST /api/broadcasts/:id/read
func (h *BroadcastHandler) MarkBroadcastAsRead(c *gin.Context) {
	broadcastIDStr := c.Param("id")
	broadcastID, err := uuid.Parse(broadcastIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid broadcast ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	if err := h.service.MarkAsRead(broadcastID, userCtx.ID); err != nil {
		log.Printf("❌ MARK_BROADCAST_READ: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark broadcast as read"})
		return
	}

	log.Printf("✅ MARK_BROADCAST_READ: Successfully marked broadcast %s as read for user %s", broadcastID, userCtx.ID)

	c.JSON(http.StatusOK, gin.H{"message": "Broadcast marked as read"})
}

// GetBroadcastStats handles GET /api/broadcasts/stats
func (h *BroadcastHandler) GetBroadcastStats(c *gin.Context) {
	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get user roles
	userService := models.NewUserService(h.service.GetDB())
	roles, err := userService.GetUserRoles(userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	// Convert roles to string slice
	var roleNames []string
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	// Get stats
	stats, err := h.service.GetStats(userCtx.ID, roleNames)
	if err != nil {
		log.Printf("❌ GET_BROADCAST_STATS: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get broadcast stats"})
		return
	}

	log.Printf("✅ GET_BROADCAST_STATS: Successfully fetched stats for user %s", userCtx.ID)

	c.JSON(http.StatusOK, stats)
}
