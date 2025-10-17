package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// ProjectStatus represents the status of a project
type ProjectStatus string

const (
	ProjectStatusDraft      ProjectStatus = "draft"
	ProjectStatusRecruiting ProjectStatus = "recruiting"
	ProjectStatusActive     ProjectStatus = "active"
	ProjectStatusCompleted  ProjectStatus = "completed"
	ProjectStatusArchived   ProjectStatus = "archived"
)

// TeamMemberStatus represents the status of a team member
type TeamMemberStatus string

const (
	TeamMemberStatusInvited   TeamMemberStatus = "invited"
	TeamMemberStatusActive    TeamMemberStatus = "active"
	TeamMemberStatusCompleted TeamMemberStatus = "completed"
	TeamMemberStatusRemoved   TeamMemberStatus = "removed"
)

// Project represents a project (formerly initiative)
type Project struct {
	ID                uuid.UUID              `json:"id" db:"id"`
	Title             string                 `json:"title" db:"title"`
	Description       string                 `json:"description" db:"description"`
	ContentJSON       map[string]interface{} `json:"content_json,omitempty" db:"content_json"`
	RequiredSkills    []string               `json:"required_skills" db:"required_skills"`
	LocationLat       *float64               `json:"location_lat" db:"location_lat"`
	LocationLng       *float64               `json:"location_lng" db:"location_lng"`
	LocationAddress   string                 `json:"location_address" db:"location_address"`
	StartDate         *time.Time             `json:"start_date" db:"start_date"`
	EndDate           *time.Time             `json:"end_date" db:"end_date"`
	Status            string                 `json:"status" db:"status"`
	ProjectStatus     ProjectStatus          `json:"project_status" db:"project_status"`
	CreatedByAdminID  uuid.UUID              `json:"created_by_admin_id" db:"created_by_admin_id"`
	TeamLeadID        *uuid.UUID             `json:"team_lead_id" db:"team_lead_id"`
	BudgetTotal       *float64               `json:"budget_total,omitempty" db:"budget_total"`
	BudgetSpent       *float64               `json:"budget_spent,omitempty" db:"budget_spent"`
	Permissions       map[string]interface{} `json:"permissions,omitempty" db:"permissions"`
	AutoNotifyMatches bool                   `json:"auto_notify_matches" db:"auto_notify_matches"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// ProjectTeamMember represents a team member in a project
type ProjectTeamMember struct {
	ID          uuid.UUID        `json:"id" db:"id"`
	ProjectID   uuid.UUID        `json:"project_id" db:"project_id"`
	VolunteerID uuid.UUID        `json:"volunteer_id" db:"volunteer_id"`
	JoinedAt    time.Time        `json:"joined_at" db:"joined_at"`
	Status      TeamMemberStatus `json:"status" db:"status"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at" db:"updated_at"`
}

// ProjectWithDetails represents a project with additional details
type ProjectWithDetails struct {
	Project
	TeamLead        *User               `json:"team_lead,omitempty"`
	CreatedByAdmin  *Admin              `json:"created_by_admin,omitempty"`
	TeamMembers     []ProjectTeamMember `json:"team_members,omitempty"`
	Applications    []Application       `json:"applications,omitempty"`
	SignupCount     int                 `json:"signup_count"`
	ActiveTeamCount int                 `json:"active_team_count"`
}

// ProjectService handles project operations
type ProjectService struct {
	db *sql.DB
}

// NewProjectService creates a new project service
func NewProjectService(db *sql.DB) *ProjectService {
	return &ProjectService{db: db}
}

// Create creates a new project
func (s *ProjectService) Create(project *Project) error {
	project.ID = uuid.New()
	skillsJSON, err := ToJSONArray(project.RequiredSkills)
	if err != nil {
		return err
	}

	return s.db.QueryRow(projectCreateQuery, project.ID, project.Title, project.Description, skillsJSON,
		project.LocationLat, project.LocationLng, project.LocationAddress, project.StartDate,
		project.EndDate, project.Status, project.ProjectStatus, project.CreatedByAdminID,
		project.TeamLeadID, project.AutoNotifyMatches).Scan(&project.CreatedAt, &project.UpdatedAt)
}

// GetByID retrieves a project by ID
func (s *ProjectService) GetByID(id uuid.UUID) (*Project, error) {
	project := &Project{}
	var skillsJSON string

	err := s.db.QueryRow(projectGetByIDQuery, id).Scan(&project.ID, &project.Title, &project.Description,
		&skillsJSON, &project.LocationLat, &project.LocationLng, &project.LocationAddress,
		&project.StartDate, &project.EndDate, &project.Status, &project.ProjectStatus,
		&project.CreatedByAdminID, &project.TeamLeadID, &project.AutoNotifyMatches, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse required skills JSON
	if err := ParseJSONArray(skillsJSON, &project.RequiredSkills); err != nil {
		return nil, err
	}

	return project, nil
}

// GetByIDWithDetails retrieves a project with full details
func (s *ProjectService) GetByIDWithDetails(id uuid.UUID) (*ProjectWithDetails, error) {
	project, err := s.GetByID(id)
	if err != nil || project == nil {
		return nil, err
	}

	details := &ProjectWithDetails{
		Project: *project,
	}

	// Get team lead details if present
	if project.TeamLeadID != nil {
		userService := NewUserService(s.db)
		teamLead, err := userService.GetByID(*project.TeamLeadID)
		if err == nil && teamLead != nil {
			details.TeamLead = &User{
				ID:    teamLead.ID,
				Email: teamLead.Email,
			}
		}
	}

	// Get created by admin details
	adminService := NewAdminService(s.db)
	admin, err := adminService.GetByID(project.CreatedByAdminID)
	if err == nil && admin != nil {
		details.CreatedByAdmin = admin
	}

	// Get team members
	teamMembers, err := s.GetProjectTeamMembers(id)
	if err == nil {
		details.TeamMembers = teamMembers
		// Count active team members
		for _, member := range teamMembers {
			if member.Status == TeamMemberStatusActive {
				details.ActiveTeamCount++
			}
		}
	}

	// Get applications count
	applicationService := NewApplicationService(s.db)
	applications, err := applicationService.GetApplicationsByProject(id)
	if err == nil {
		details.Applications = applications
		details.SignupCount = len(applications)
	}

	return details, nil
}

// List retrieves projects with optional filtering
func (s *ProjectService) List(limit, offset int, status *string, skills []string) ([]Project, error) {
	// Build query based on whether skills filter is provided
	var query string
	var args []interface{}

	if len(skills) > 0 {
		// Query with skills filter
		query = projectListWithSkillsQuery
		args = []interface{}{status, skills, limit, offset}
	} else {
		// Query without skills filter
		query = projectListQuery
		args = []interface{}{status, limit, offset}
	}

	log.Printf("🔍 PROJECT_LIST_QUERY: Executing query with params - status=%v, skills=%v, limit=%d, offset=%d", status, skills, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		log.Printf("❌ PROJECT_LIST_QUERY: Database query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	rowCount := 0
	for rows.Next() {
		rowCount++
		var project Project
		var skillsJSON string
		err := rows.Scan(&project.ID, &project.Title, &project.Description,
			&project.LocationLat, &project.LocationLng, &project.LocationAddress,
			&project.StartDate, &project.EndDate, &project.Status, &project.ProjectStatus,
			&project.CreatedByAdminID, &project.TeamLeadID, &project.AutoNotifyMatches, &project.CreatedAt, &project.UpdatedAt,
			&skillsJSON)
		if err != nil {
			log.Printf("❌ PROJECT_LIST_SCAN: Row %d scan failed: %v", rowCount, err)
			return nil, err
		}

		// Parse required skills JSON - now contains objects with id and name
		var skillObjects []map[string]interface{}
		if err := json.Unmarshal([]byte(skillsJSON), &skillObjects); err != nil {
			log.Printf("❌ PROJECT_LIST_PARSE: Failed to parse skills for project %s: %v", project.ID, err)
			return nil, err
		}

		// Convert skill objects to skill names
		var skillNames []string
		for _, skillObj := range skillObjects {
			if name, ok := skillObj["name"].(string); ok {
				skillNames = append(skillNames, name)
			}
		}
		project.RequiredSkills = skillNames

		projects = append(projects, project)
	}

	log.Printf("✅ PROJECT_LIST_COMPLETE: Successfully processed %d projects", len(projects))
	return projects, nil
}

// ListByTeamLead retrieves projects for a specific team lead
func (s *ProjectService) ListByTeamLead(teamLeadID uuid.UUID, limit, offset int) ([]Project, error) {
	rows, err := s.db.Query(projectListByTeamLeadQuery, teamLeadID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var project Project
		var skillsJSON string
		err := rows.Scan(&project.ID, &project.Title, &project.Description,
			&project.LocationLat, &project.LocationLng, &project.LocationAddress,
			&project.StartDate, &project.EndDate, &project.Status, &project.ProjectStatus,
			&project.CreatedByAdminID, &project.TeamLeadID, &project.AutoNotifyMatches, &project.CreatedAt, &project.UpdatedAt,
			&skillsJSON)
		if err != nil {
			return nil, err
		}

		// Parse required skills JSON - now contains objects with id and name
		var skillObjects []map[string]interface{}
		if err := json.Unmarshal([]byte(skillsJSON), &skillObjects); err != nil {
			return nil, err
		}

		// Convert skill objects to skill names
		var skillNames []string
		for _, skillObj := range skillObjects {
			if name, ok := skillObj["name"].(string); ok {
				skillNames = append(skillNames, name)
			}
		}
		project.RequiredSkills = skillNames

		projects = append(projects, project)
	}

	return projects, nil
}

// Update updates a project
func (s *ProjectService) Update(project *Project) error {
	skillsJSON, err := ToJSONArray(project.RequiredSkills)
	if err != nil {
		return err
	}

	return s.db.QueryRow(projectUpdateQuery, project.ID, project.Title, project.Description, skillsJSON,
		project.LocationLat, project.LocationLng, project.LocationAddress, project.StartDate,
		project.EndDate, project.Status, project.ProjectStatus, project.TeamLeadID, project.AutoNotifyMatches).
		Scan(&project.UpdatedAt)
}

// Delete deletes a project
func (s *ProjectService) Delete(id uuid.UUID) error {
	_, err := s.db.Exec(projectDeleteQuery, id)
	return err
}

// GetProjectTeamMembers retrieves team members for a project
func (s *ProjectService) GetProjectTeamMembers(projectID uuid.UUID) ([]ProjectTeamMember, error) {
	rows, err := s.db.Query(projectGetTeamMembersQuery, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []ProjectTeamMember
	for rows.Next() {
		var member ProjectTeamMember
		err := rows.Scan(&member.ID, &member.ProjectID, &member.VolunteerID,
			&member.JoinedAt, &member.Status, &member.CreatedAt, &member.UpdatedAt)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, nil
}

// AddTeamMember adds a volunteer to a project team
func (s *ProjectService) AddTeamMember(projectID, volunteerID uuid.UUID, status TeamMemberStatus) error {
	_, err := s.db.Exec(projectAddTeamMemberQuery, uuid.New(), projectID, volunteerID, status)
	return err
}

// UpdateTeamMemberStatus updates a team member's status
func (s *ProjectService) UpdateTeamMemberStatus(projectID, volunteerID uuid.UUID, status TeamMemberStatus) error {
	result, err := s.db.Exec(projectUpdateTeamMemberStatusQuery, projectID, volunteerID, status)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows // Team member not found
	}

	return nil
}

// RemoveTeamMember removes a volunteer from a project team
func (s *ProjectService) RemoveTeamMember(projectID, volunteerID uuid.UUID) error {
	_, err := s.db.Exec(projectRemoveTeamMemberQuery, projectID, volunteerID)
	return err
}

// GetProjectSignups retrieves all applications/signups for a project
func (s *ProjectService) GetProjectSignups(projectID uuid.UUID) ([]Application, error) {
	applicationService := NewApplicationService(s.db)
	return applicationService.GetApplicationsByProject(projectID)
}

// IsTeamLead checks if a user is the team lead for a project
func (s *ProjectService) IsTeamLead(projectID, userID uuid.UUID) (bool, error) {
	var count int
	err := s.db.QueryRow(projectIsTeamLeadQuery, projectID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// IsTeamMember checks if a user is an active team member of a project
func (s *ProjectService) IsTeamMember(projectID, userID uuid.UUID) (bool, error) {
	var count int
	err := s.db.QueryRow(projectIsTeamMemberQuery, projectID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// AssignTeamLead assigns a team lead to a project
func (s *ProjectService) AssignTeamLead(projectID, teamLeadID uuid.UUID) error {
	result, err := s.db.Exec(projectAssignTeamLeadQuery, projectID, teamLeadID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows // Project not found
	}

	return nil
}

// GetActiveProjects retrieves projects that are currently active
func (s *ProjectService) GetActiveProjects() ([]Project, error) {
	rows, err := s.db.Query(projectGetActiveProjectsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var project Project
		var skillsJSON string
		err := rows.Scan(&project.ID, &project.Title, &project.Description,
			&project.LocationLat, &project.LocationLng, &project.LocationAddress,
			&project.StartDate, &project.EndDate, &project.Status, &project.ProjectStatus,
			&project.CreatedByAdminID, &project.TeamLeadID, &project.AutoNotifyMatches, &project.CreatedAt, &project.UpdatedAt,
			&skillsJSON)
		if err != nil {
			return nil, err
		}

		// Parse required skills JSON - now contains objects with id and name
		var skillObjects []map[string]interface{}
		if err := json.Unmarshal([]byte(skillsJSON), &skillObjects); err != nil {
			return nil, err
		}

		// Convert skill objects to skill names
		var skillNames []string
		for _, skillObj := range skillObjects {
			if name, ok := skillObj["name"].(string); ok {
				skillNames = append(skillNames, name)
			}
		}
		project.RequiredSkills = skillNames

		projects = append(projects, project)
	}

	return projects, nil
}

// GetDB returns the database connection (needed for creating other services)
func (s *ProjectService) GetDB() *sql.DB {
	return s.db
}

// IsProjectCreator checks if a user is the creator of a project
func (s *ProjectService) IsProjectCreator(projectID, userID uuid.UUID) (bool, error) {
	var count int
	err := s.db.QueryRow(projectIsCreatorQuery, projectID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CanEditProject checks if a user can edit a project (team lead, admin, or creator)
func (s *ProjectService) CanEditProject(projectID, userID uuid.UUID) (bool, error) {
	// Check if user is admin
	userService := NewUserService(s.db)
	roles, err := userService.GetUserRoles(userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.Name == "admin" {
			return true, nil
		}
	}

	// Check if user is team lead
	isTeamLead, err := s.IsTeamLead(projectID, userID)
	if err != nil {
		return false, err
	}
	if isTeamLead {
		return true, nil
	}

	// Check if user is creator
	isCreator, err := s.IsProjectCreator(projectID, userID)
	if err != nil {
		return false, err
	}

	return isCreator, nil
}

// TransitionProjectStatus transitions a project to a new status with validation
func (s *ProjectService) TransitionProjectStatus(projectID uuid.UUID, newStatus ProjectStatus, userID uuid.UUID) error {
	// Get current project
	project, err := s.GetByID(projectID)
	if err != nil {
		return err
	}
	if project == nil {
		return sql.ErrNoRows
	}

	// Check if user can edit project
	canEdit, err := s.CanEditProject(projectID, userID)
	if err != nil {
		return err
	}
	if !canEdit {
		return sql.ErrNoRows // Permission denied
	}

	// Validate transition
	if err := s.validateStatusTransition(project.ProjectStatus, newStatus, projectID); err != nil {
		return err
	}

	// Update project status
	_, err = s.db.Exec(projectTransitionStatusQuery, projectID, newStatus)
	return err
}

// validateStatusTransition validates if a status transition is allowed
func (s *ProjectService) validateStatusTransition(currentStatus, newStatus ProjectStatus, projectID uuid.UUID) error {
	// Same status is always allowed
	if currentStatus == newStatus {
		return nil
	}

	// Define valid transitions
	validTransitions := map[ProjectStatus][]ProjectStatus{
		ProjectStatusDraft:      {ProjectStatusRecruiting, ProjectStatusArchived},
		ProjectStatusRecruiting: {ProjectStatusActive, ProjectStatusArchived},
		ProjectStatusActive:     {ProjectStatusCompleted, ProjectStatusArchived},
		ProjectStatusCompleted:  {ProjectStatusArchived},
		ProjectStatusArchived:   {}, // No transitions from archived
	}

	// Check if transition is valid
	allowedTransitions, exists := validTransitions[currentStatus]
	if !exists {
		return fmt.Errorf("invalid current status: %s", currentStatus)
	}

	validTransition := false
	for _, allowed := range allowedTransitions {
		if allowed == newStatus {
			validTransition = true
			break
		}
	}

	if !validTransition {
		return fmt.Errorf("invalid transition from %s to %s", currentStatus, newStatus)
	}

	// Get project details for validation
	project, err := s.GetByID(projectID)
	if err != nil {
		return fmt.Errorf("failed to get project details: %w", err)
	}

	// Validate specific transition requirements
	switch newStatus {
	case ProjectStatusRecruiting:
		// Must have team lead assigned
		if project.TeamLeadID == nil {
			return fmt.Errorf("cannot transition to recruiting: team lead must be assigned")
		}
	case ProjectStatusActive:
		// Must have at least one active team member
		activeCount, err := s.getActiveTeamMemberCount(projectID)
		if err != nil {
			return err
		}
		if activeCount == 0 {
			return fmt.Errorf("cannot transition to active: must have at least one active team member")
		}
	}

	return nil
}

// getActiveTeamMemberCount returns the count of active team members for a project
func (s *ProjectService) getActiveTeamMemberCount(projectID uuid.UUID) (int, error) {
	var count int
	err := s.db.QueryRow(projectActiveTeamCountQuery, projectID).Scan(&count)
	return count, err
}
