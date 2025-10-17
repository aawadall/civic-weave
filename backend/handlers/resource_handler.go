package handlers

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"civicweave/backend/middleware"
	"civicweave/backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ResourceHandler handles resource-related requests
type ResourceHandler struct {
	service *models.ResourceService
}

// NewResourceHandler creates a new resource handler
func NewResourceHandler(service *models.ResourceService) *ResourceHandler {
	return &ResourceHandler{
		service: service,
	}
}

// CreateResourceRequest represents resource creation request
type CreateResourceRequest struct {
	Title        string   `json:"title" binding:"required"`
	Description  string   `json:"description"`
	ResourceType string   `json:"resource_type" binding:"required"`
	Scope        string   `json:"scope" binding:"required"`
	ProjectID    *string  `json:"project_id,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	FileURL      string   `json:"file_url,omitempty"` // For links
}

// UpdateResourceRequest represents resource update request
type UpdateResourceRequest struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	ResourceType string   `json:"resource_type"`
	Scope        string   `json:"scope"`
	ProjectID    *string  `json:"project_id,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	FileURL      string   `json:"file_url,omitempty"`
}

// ListResources handles GET /api/resources
func (h *ResourceHandler) ListResources(c *gin.Context) {
	// Get query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	scope := c.Query("scope")
	projectIDStr := c.Query("project_id")
	search := c.Query("search")
	tags := c.QueryArray("tags")
	resourceType := c.Query("type")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Build filters
	filters := models.ResourceFilters{
		Search: &search,
		Tags:   tags,
		Type:   &resourceType,
	}

	if scope != "" {
		filters.Scope = &scope
	}

	if projectIDStr != "" {
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			return
		}
		filters.ProjectID = &projectID
	}

	// Get resources
	resources, err := h.service.List(filters, limit, offset)
	if err != nil {
		log.Printf("❌ LIST_RESOURCES: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get resources"})
		return
	}

	log.Printf("✅ LIST_RESOURCES: Successfully fetched %d resources", len(resources))

	c.JSON(http.StatusOK, gin.H{
		"resources": resources,
		"limit":     limit,
		"offset":    offset,
		"count":     len(resources),
	})
}

// GetResource handles GET /api/resources/:id
func (h *ResourceHandler) GetResource(c *gin.Context) {
	resourceIDStr := c.Param("id")
	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
		return
	}

	// Get resource
	resource, err := h.service.GetByID(resourceID)
	if err != nil {
		log.Printf("❌ GET_RESOURCE: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get resource"})
		return
	}

	if resource == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
		return
	}

	log.Printf("✅ GET_RESOURCE: Successfully fetched resource %s", resourceID)

	c.JSON(http.StatusOK, resource)
}

// CreateResource handles POST /api/resources
func (h *ResourceHandler) CreateResource(c *gin.Context) {
	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has permission (team_lead or admin)
	userService := models.NewUserService(h.service.GetDB())
	roles, err := userService.GetUserRoles(userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	hasPermission := false
	for _, role := range roles {
		if role.Name == "admin" || role.Name == "team_lead" {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team leads and admins can create resources"})
		return
	}

	// Parse form data
	title := c.PostForm("title")
	description := c.PostForm("description")
	resourceType := c.PostForm("resource_type")
	scope := c.PostForm("scope")
	projectIDStr := c.PostForm("project_id")
	tagsStr := c.PostForm("tags")

	if title == "" || resourceType == "" || scope == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title, resource_type, and scope are required"})
		return
	}

	// Parse project ID if provided
	var projectID *uuid.UUID
	if projectIDStr != "" {
		parsedID, err := uuid.Parse(projectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			return
		}
		projectID = &parsedID
	}

	// Parse tags
	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
	}

	// Handle file upload or link
	var fileURL string
	var fileSize *int64
	var mimeType *string

	if resourceType == "file" {
		// Handle file upload
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File upload required for file type"})
			return
		}
		defer file.Close()

		// Create uploads directory if it doesn't exist
		uploadDir := "uploads"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			log.Printf("❌ CREATE_RESOURCE: Failed to create upload directory: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
			return
		}

		// Generate unique filename
		ext := filepath.Ext(header.Filename)
		filename := uuid.New().String() + ext
		filePath := filepath.Join(uploadDir, filename)

		// Save file
		dst, err := os.Create(filePath)
		if err != nil {
			log.Printf("❌ CREATE_RESOURCE: Failed to create file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			log.Printf("❌ CREATE_RESOURCE: Failed to copy file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		fileURL = "/uploads/" + filename
		fileSizeInt := header.Size
		fileSize = &fileSizeInt
		mimeTypeStr := header.Header.Get("Content-Type")
		mimeType = &mimeTypeStr
	} else if resourceType == "link" {
		// Handle link
		fileURL = c.PostForm("file_url")
		if fileURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file_url is required for link type"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource type"})
		return
	}

	// Create resource
	resource := &models.Resource{
		Title:        title,
		Description:  description,
		ResourceType: resourceType,
		FileURL:      fileURL,
		FileSize:     fileSize,
		MimeType:     mimeType,
		Scope:        scope,
		ProjectID:    projectID,
		UploadedByID: userCtx.ID,
		Tags:         tags,
	}

	if err := h.service.Create(resource); err != nil {
		log.Printf("❌ CREATE_RESOURCE: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create resource"})
		return
	}

	log.Printf("✅ CREATE_RESOURCE: Successfully created resource %s", resource.ID)

	c.JSON(http.StatusCreated, resource)
}

// UpdateResource handles PUT /api/resources/:id
func (h *ResourceHandler) UpdateResource(c *gin.Context) {
	resourceIDStr := c.Param("id")
	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
		return
	}

	var req UpdateResourceRequest
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

	// Check if user is uploader or admin
	isUploader, err := h.service.IsUploader(resourceID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check uploader status"})
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

	if !isUploader && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only resource uploader or admin can update"})
		return
	}

	// Get existing resource
	resource, err := h.service.GetByID(resourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get resource"})
		return
	}

	if resource == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
		return
	}

	// Update fields
	if req.Title != "" {
		resource.Title = req.Title
	}
	if req.Description != "" {
		resource.Description = req.Description
	}
	if req.ResourceType != "" {
		resource.ResourceType = req.ResourceType
	}
	if req.Scope != "" {
		resource.Scope = req.Scope
	}
	if req.FileURL != "" {
		resource.FileURL = req.FileURL
	}
	if req.Tags != nil {
		resource.Tags = req.Tags
	}
	if req.ProjectID != nil {
		if *req.ProjectID == "" {
			resource.ProjectID = nil
		} else {
			projectID, err := uuid.Parse(*req.ProjectID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
				return
			}
			resource.ProjectID = &projectID
		}
	}

	if err := h.service.Update(resource); err != nil {
		log.Printf("❌ UPDATE_RESOURCE: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update resource"})
		return
	}

	log.Printf("✅ UPDATE_RESOURCE: Successfully updated resource %s", resourceID)

	c.JSON(http.StatusOK, resource)
}

// DeleteResource handles DELETE /api/resources/:id
func (h *ResourceHandler) DeleteResource(c *gin.Context) {
	resourceIDStr := c.Param("id")
	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user is uploader or admin
	isUploader, err := h.service.IsUploader(resourceID, userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check uploader status"})
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

	if !isUploader && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only resource uploader or admin can delete"})
		return
	}

	if err := h.service.SoftDelete(resourceID); err != nil {
		log.Printf("❌ DELETE_RESOURCE: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete resource"})
		return
	}

	log.Printf("✅ DELETE_RESOURCE: Successfully deleted resource %s", resourceID)

	c.JSON(http.StatusOK, gin.H{"message": "Resource deleted successfully"})
}

// DownloadResource handles GET /api/resources/:id/download
func (h *ResourceHandler) DownloadResource(c *gin.Context) {
	resourceIDStr := c.Param("id")
	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
		return
	}

	// Get resource
	resource, err := h.service.GetByID(resourceID)
	if err != nil {
		log.Printf("❌ DOWNLOAD_RESOURCE: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get resource"})
		return
	}

	if resource == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
		return
	}

	// Increment download count
	if err := h.service.IncrementDownloadCount(resourceID); err != nil {
		log.Printf("❌ DOWNLOAD_RESOURCE: Failed to increment download count: %v", err)
		// Don't fail the request for this
	}

	// Handle different resource types
	if resource.ResourceType == "file" {
		// Serve file
		filePath := strings.TrimPrefix(resource.FileURL, "/uploads/")
		fullPath := filepath.Join("uploads", filePath)

		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found on disk"})
			return
		}

		c.Header("Content-Disposition", "attachment; filename="+resource.Title)
		if resource.MimeType != nil {
			c.Header("Content-Type", *resource.MimeType)
		}
		c.File(fullPath)
	} else if resource.ResourceType == "link" {
		// Redirect to external URL
		c.Redirect(http.StatusFound, resource.FileURL)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource type for download"})
		return
	}

	log.Printf("✅ DOWNLOAD_RESOURCE: Successfully served resource %s", resourceID)
}

// GetResourceStats handles GET /api/resources/stats
func (h *ResourceHandler) GetResourceStats(c *gin.Context) {
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can view resource stats"})
		return
	}

	// Get stats
	stats, err := h.service.GetStats()
	if err != nil {
		log.Printf("❌ GET_RESOURCE_STATS: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get resource stats"})
		return
	}

	log.Printf("✅ GET_RESOURCE_STATS: Successfully fetched stats for user %s", userCtx.ID)

	c.JSON(http.StatusOK, stats)
}

// GetRecentResources handles GET /api/resources/recent
func (h *ResourceHandler) GetRecentResources(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	// Get recent resources
	resources, err := h.service.GetRecentResources(limit)
	if err != nil {
		// Check if the error is due to missing table
		if strings.Contains(err.Error(), "does not exist") {
			log.Printf("⚠️ GET_RECENT_RESOURCES: Resources table not found, returning empty list")
			c.JSON(http.StatusOK, gin.H{
				"resources": []interface{}{},
				"count":     0,
			})
			return
		}
		log.Printf("❌ GET_RECENT_RESOURCES: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recent resources"})
		return
	}

	log.Printf("✅ GET_RECENT_RESOURCES: Successfully fetched %d recent resources", len(resources))

	c.JSON(http.StatusOK, gin.H{
		"resources": resources,
		"count":     len(resources),
	})
}
