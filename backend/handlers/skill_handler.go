package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"civicweave/backend/models"
)

// SkillHandler handles skill-related API endpoints
type SkillHandler struct {
	taxonomyService *models.SkillTaxonomyService
}

// NewSkillHandler creates a new skill handler
func NewSkillHandler(taxonomyService *models.SkillTaxonomyService) *SkillHandler {
	return &SkillHandler{
		taxonomyService: taxonomyService,
	}
}

// GetTaxonomy handles GET /api/skills/taxonomy
func (h *SkillHandler) GetTaxonomy(c *gin.Context) {
	skills, err := h.taxonomyService.GetAllSkills()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve skills taxonomy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"skills": skills,
		"count":  len(skills),
	})
}

// AddSkill handles POST /api/skills/taxonomy
func (h *SkillHandler) AddSkill(c *gin.Context) {
	type Request struct {
		SkillName string `json:"skill_name" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.SkillName) < 2 || len(req.SkillName) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Skill name must be between 2 and 100 characters"})
		return
	}

	skill, err := h.taxonomyService.AddSkill(req.SkillName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add skill"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"skill": skill,
		"message": "Skill added successfully",
	})
}

// GetVolunteerSkills handles GET /api/volunteers/me/skills
func (h *SkillHandler) GetVolunteerSkills(c *gin.Context) {
	volunteerID, exists := c.Get("volunteer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Volunteer ID not found"})
		return
	}

	volunteerUUID, ok := volunteerID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid volunteer ID"})
		return
	}

	skills, err := h.taxonomyService.GetVolunteerSkills(volunteerUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve volunteer skills"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"skills": skills,
		"count":  len(skills),
	})
}

// UpdateVolunteerSkills handles PUT /api/volunteers/me/skills
func (h *SkillHandler) UpdateVolunteerSkills(c *gin.Context) {
	volunteerID, exists := c.Get("volunteer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Volunteer ID not found"})
		return
	}

	volunteerUUID, ok := volunteerID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid volunteer ID"})
		return
	}

	type Request struct {
		SkillNames []string `json:"skill_names" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.SkillNames) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 50 skills allowed"})
		return
	}

	// Validate skill names
	for _, name := range req.SkillNames {
		if len(name) < 2 || len(name) > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Each skill name must be between 2 and 100 characters"})
			return
		}
	}

	// Resolve skill names to IDs (adds new skills to taxonomy if needed)
	skillIDs, err := h.taxonomyService.ResolveSkillNames(req.SkillNames)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve skill names"})
		return
	}

	// Update volunteer skills
	err = h.taxonomyService.UpdateVolunteerSkills(volunteerUUID, skillIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update volunteer skills"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Skills updated successfully",
		"skills_added": len(skillIDs),
	})
}

// AddVolunteerSkills handles POST /api/volunteers/me/skills
func (h *SkillHandler) AddVolunteerSkills(c *gin.Context) {
	volunteerID, exists := c.Get("volunteer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Volunteer ID not found"})
		return
	}

	volunteerUUID, ok := volunteerID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid volunteer ID"})
		return
	}

	type Request struct {
		SkillNames []string `json:"skill_names" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.SkillNames) > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 20 skills allowed per request"})
		return
	}

	// Validate skill names
	for _, name := range req.SkillNames {
		if len(name) < 2 || len(name) > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Each skill name must be between 2 and 100 characters"})
			return
		}
	}

	// Resolve skill names to IDs
	skillIDs, err := h.taxonomyService.ResolveSkillNames(req.SkillNames)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve skill names"})
		return
	}

	// Add skills to volunteer (doesn't remove existing)
	err = h.taxonomyService.AddVolunteerSkills(volunteerUUID, skillIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add volunteer skills"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Skills added successfully",
		"skills_added": len(skillIDs),
	})
}

// RemoveVolunteerSkill handles DELETE /api/volunteers/me/skills/:skill_id
func (h *SkillHandler) RemoveVolunteerSkill(c *gin.Context) {
	volunteerID, exists := c.Get("volunteer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Volunteer ID not found"})
		return
	}

	volunteerUUID, ok := volunteerID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid volunteer ID"})
		return
	}

	skillIDStr := c.Param("skill_id")
	skillID, err := strconv.Atoi(skillIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid skill ID"})
		return
	}

	err = h.taxonomyService.RemoveVolunteerSkill(volunteerUUID, skillID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove volunteer skill"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Skill removed successfully",
	})
}

// GetProfileCompletion handles GET /api/volunteers/me/profile-completion
func (h *SkillHandler) GetProfileCompletion(c *gin.Context) {
	volunteerID, exists := c.Get("volunteer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Volunteer ID not found"})
		return
	}

	volunteerUUID, ok := volunteerID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid volunteer ID"})
		return
	}

	completion, err := h.taxonomyService.GetVolunteerProfileCompletion(volunteerUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate profile completion"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"completion_percentage": completion,
		"is_complete": completion >= 100,
	})
}

// GetInitiativeSkills handles GET /api/initiatives/:id/skills
func (h *SkillHandler) GetInitiativeSkills(c *gin.Context) {
	initiativeIDStr := c.Param("id")
	initiativeID, err := uuid.Parse(initiativeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid initiative ID"})
		return
	}

	skills, err := h.taxonomyService.GetInitiativeSkills(initiativeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve initiative skills"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"skills": skills,
		"count":  len(skills),
	})
}

// UpdateInitiativeSkills handles PUT /api/initiatives/:id/skills
func (h *SkillHandler) UpdateInitiativeSkills(c *gin.Context) {
	initiativeIDStr := c.Param("id")
	initiativeID, err := uuid.Parse(initiativeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid initiative ID"})
		return
	}

	type Request struct {
		SkillNames []string `json:"skill_names" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.SkillNames) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 50 skills allowed"})
		return
	}

	// Validate skill names
	for _, name := range req.SkillNames {
		if len(name) < 2 || len(name) > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Each skill name must be between 2 and 100 characters"})
			return
		}
	}

	// Resolve skill names to IDs
	skillIDs, err := h.taxonomyService.ResolveSkillNames(req.SkillNames)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve skill names"})
		return
	}

	// Update initiative skills
	err = h.taxonomyService.UpdateInitiativeSkills(initiativeID, skillIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update initiative skills"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Initiative skills updated successfully",
		"skills_added": len(skillIDs),
	})
}
