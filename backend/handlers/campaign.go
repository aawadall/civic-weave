package handlers

import (
	"net/http"
	"strconv"
	"time"

	"civicweave/backend/config"
	"civicweave/backend/middleware"
	"civicweave/backend/models"
	"civicweave/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CampaignHandler handles campaign-related requests
type CampaignHandler struct {
	campaignService *models.CampaignService
	emailService    *services.EmailService
	config          *config.Config
}

// NewCampaignHandler creates a new campaign handler
func NewCampaignHandler(campaignService *models.CampaignService, emailService *services.EmailService, config *config.Config) *CampaignHandler {
	return &CampaignHandler{
		campaignService: campaignService,
		emailService:    emailService,
		config:          config,
	}
}

// ListCampaigns handles GET /api/campaigns
func (h *CampaignHandler) ListCampaigns(c *gin.Context) {
	// Get query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	statusParam := c.Query("status")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
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

	// Check if user has permission to view campaigns (campaign_manager or admin)
	if !userCtx.HasAnyRole("campaign_manager", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to view campaigns"})
		return
	}

	var status *models.CampaignStatus
	if statusParam != "" {
		s := models.CampaignStatus(statusParam)
		status = &s
	}

	var createdByUserID *uuid.UUID
	// If user is not admin, only show their own campaigns
	if !userCtx.HasRole("admin") {
		createdByUserID = &userCtx.ID
	}

	campaigns, err := h.campaignService.ListCampaigns(limit, offset, status, createdByUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get campaigns"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"campaigns": campaigns,
		"limit":     limit,
		"offset":    offset,
		"count":     len(campaigns),
	})
}

// GetCampaignByID handles GET /api/campaigns/:id
func (h *CampaignHandler) GetCampaignByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has permission to view campaigns (campaign_manager or admin)
	if !userCtx.HasAnyRole("campaign_manager", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to view campaigns"})
		return
	}

	campaign, err := h.campaignService.GetCampaignByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get campaign"})
		return
	}

	if campaign == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
		return
	}

	// If user is not admin, check if they created this campaign
	if !userCtx.HasRole("admin") && campaign.CreatedByUserID != userCtx.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to view this campaign"})
		return
	}

	c.JSON(http.StatusOK, campaign)
}

// CreateCampaign handles POST /api/campaigns
func (h *CampaignHandler) CreateCampaign(c *gin.Context) {
	var req struct {
		Title        string   `json:"title" binding:"required"`
		Description  string   `json:"description"`
		TargetRoles  []string `json:"target_roles" binding:"required"`
		EmailSubject string   `json:"email_subject" binding:"required"`
		EmailBody    string   `json:"email_body" binding:"required"`
		ScheduledAt  *string  `json:"scheduled_at"`
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

	// Check if user has permission to create campaigns (campaign_manager or admin)
	if !userCtx.HasAnyRole("campaign_manager", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to create campaigns"})
		return
	}

	// Parse scheduled_at if provided
	var scheduledAt *time.Time
	if req.ScheduledAt != nil && *req.ScheduledAt != "" {
		if parsed, err := time.Parse(time.RFC3339, *req.ScheduledAt); err == nil {
			scheduledAt = &parsed
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scheduled_at format. Use RFC3339 format."})
			return
		}
	}

	// Determine initial status
	status := models.CampaignStatusDraft
	if scheduledAt != nil {
		status = models.CampaignStatusScheduled
	}

	campaign := &models.Campaign{
		Title:           req.Title,
		Description:     req.Description,
		TargetRoles:     req.TargetRoles,
		Status:          status,
		EmailSubject:    req.EmailSubject,
		EmailBody:       req.EmailBody,
		CreatedByUserID: userCtx.ID,
		ScheduledAt:     scheduledAt,
	}

	if err := h.campaignService.CreateCampaign(campaign); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create campaign"})
		return
	}

	c.JSON(http.StatusCreated, campaign)
}

// UpdateCampaign handles PUT /api/campaigns/:id
func (h *CampaignHandler) UpdateCampaign(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
		return
	}

	var req struct {
		Title        string   `json:"title" binding:"required"`
		Description  string   `json:"description"`
		TargetRoles  []string `json:"target_roles" binding:"required"`
		Status       string   `json:"status"`
		EmailSubject string   `json:"email_subject" binding:"required"`
		EmailBody    string   `json:"email_body" binding:"required"`
		ScheduledAt  *string  `json:"scheduled_at"`
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

	// Check if user has permission to update campaigns (campaign_manager or admin)
	if !userCtx.HasAnyRole("campaign_manager", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to update campaigns"})
		return
	}

	// Get existing campaign to check ownership
	existingCampaign, err := h.campaignService.GetCampaignByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get campaign"})
		return
	}
	if existingCampaign == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
		return
	}

	// If user is not admin, check if they created this campaign
	if !userCtx.HasRole("admin") && existingCampaign.CreatedByUserID != userCtx.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to update this campaign"})
		return
	}

	// Parse scheduled_at if provided
	var scheduledAt *time.Time
	if req.ScheduledAt != nil && *req.ScheduledAt != "" {
		if parsed, err := time.Parse(time.RFC3339, *req.ScheduledAt); err == nil {
			scheduledAt = &parsed
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scheduled_at format. Use RFC3339 format."})
			return
		}
	}

	// Validate status
	status := models.CampaignStatus(req.Status)
	if status != "" && status != models.CampaignStatusDraft && status != models.CampaignStatusScheduled &&
		status != models.CampaignStatusSending && status != models.CampaignStatusSent && status != models.CampaignStatusFailed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign status"})
		return
	}

	campaign := &models.Campaign{
		ID:              id,
		Title:           req.Title,
		Description:     req.Description,
		TargetRoles:     req.TargetRoles,
		Status:          status,
		EmailSubject:    req.EmailSubject,
		EmailBody:       req.EmailBody,
		CreatedByUserID: existingCampaign.CreatedByUserID,
		ScheduledAt:     scheduledAt,
		SentAt:          existingCampaign.SentAt,
		CreatedAt:       existingCampaign.CreatedAt,
		UpdatedAt:       existingCampaign.UpdatedAt,
	}

	if err := h.campaignService.UpdateCampaign(campaign); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update campaign"})
		return
	}

	c.JSON(http.StatusOK, campaign)
}

// DeleteCampaign handles DELETE /api/campaigns/:id
func (h *CampaignHandler) DeleteCampaign(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has permission to delete campaigns (admin only for now)
	if !userCtx.HasRole("admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to delete campaigns"})
		return
	}

	if err := h.campaignService.DeleteCampaign(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete campaign"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetCampaignStats handles GET /api/campaigns/:id/stats
func (h *CampaignHandler) GetCampaignStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has permission to view campaign stats (campaign_manager or admin)
	if !userCtx.HasAnyRole("campaign_manager", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to view campaign stats"})
		return
	}

	// Get campaign to check ownership
	campaign, err := h.campaignService.GetCampaignByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get campaign"})
		return
	}
	if campaign == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
		return
	}

	// If user is not admin, check if they created this campaign
	if !userCtx.HasRole("admin") && campaign.CreatedByUserID != userCtx.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to view this campaign's stats"})
		return
	}

	stats, err := h.campaignService.GetCampaignStats(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get campaign stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetCampaignRecipients handles GET /api/campaigns/:id/recipients
func (h *CampaignHandler) GetCampaignRecipients(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has permission to view campaign recipients (campaign_manager or admin)
	if !userCtx.HasAnyRole("campaign_manager", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to view campaign recipients"})
		return
	}

	// Get campaign to check ownership
	campaign, err := h.campaignService.GetCampaignByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get campaign"})
		return
	}
	if campaign == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
		return
	}

	// If user is not admin, check if they created this campaign
	if !userCtx.HasRole("admin") && campaign.CreatedByUserID != userCtx.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to view this campaign's recipients"})
		return
	}

	recipients, err := h.campaignService.GetCampaignRecipients(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get campaign recipients"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"recipients": recipients})
}

// PreviewCampaign handles GET /api/campaigns/:id/preview
func (h *CampaignHandler) PreviewCampaign(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has permission to preview campaigns (campaign_manager or admin)
	if !userCtx.HasAnyRole("campaign_manager", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to preview campaigns"})
		return
	}

	// Get campaign to check ownership
	campaign, err := h.campaignService.GetCampaignByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get campaign"})
		return
	}
	if campaign == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
		return
	}

	// If user is not admin, check if they created this campaign
	if !userCtx.HasRole("admin") && campaign.CreatedByUserID != userCtx.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to preview this campaign"})
		return
	}

	// Get target users for preview
	targetUsers, err := h.campaignService.GetTargetUsersForCampaign(campaign.TargetRoles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get target users for preview"})
		return
	}

	preview := gin.H{
		"campaign":     campaign,
		"target_count": len(targetUsers),
		"target_users": targetUsers,
	}

	c.JSON(http.StatusOK, preview)
}

// SendCampaign handles POST /api/campaigns/:id/send
func (h *CampaignHandler) SendCampaign(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
		return
	}

	// Get user context
	userCtx, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has permission to send campaigns (campaign_manager or admin)
	if !userCtx.HasAnyRole("campaign_manager", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to send campaigns"})
		return
	}

	// Get campaign to check ownership and status
	campaign, err := h.campaignService.GetCampaignByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get campaign"})
		return
	}
	if campaign == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
		return
	}

	// If user is not admin, check if they created this campaign
	if !userCtx.HasRole("admin") && campaign.CreatedByUserID != userCtx.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to send this campaign"})
		return
	}

	// Check if campaign can be sent
	if campaign.Status != models.CampaignStatusDraft && campaign.Status != models.CampaignStatusScheduled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Campaign cannot be sent in its current status"})
		return
	}

	// Get target users for the campaign
	targetUsers, err := h.campaignService.GetTargetUsersForCampaign(campaign.TargetRoles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get target users for campaign"})
		return
	}

	if len(targetUsers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No target users found for the specified roles"})
		return
	}

	// Extract email addresses
	var recipients []string
	for _, user := range targetUsers {
		recipients = append(recipients, user.Email)
	}

	// Send campaign emails
	err = h.emailService.SendCampaignEmail(recipients, campaign.EmailSubject, campaign.EmailBody, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send campaign emails"})
		return
	}

	// Update campaign status to sent
	campaign.Status = models.CampaignStatusSent
	now := time.Now()
	campaign.SentAt = &now

	if err := h.campaignService.UpdateCampaign(campaign); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update campaign status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Campaign sent successfully",
		"recipients":  len(recipients),
		"campaign_id": campaign.ID,
	})
}
