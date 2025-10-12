package models

import (
	"database/sql"
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
	ID               uuid.UUID              `json:"id" db:"id"`
	Title            string                 `json:"title" db:"title"`
	Description      string                 `json:"description" db:"description"`
	ContentJSON      map[string]interface{} `json:"content_json,omitempty" db:"content_json"`
	RequiredSkills   []string               `json:"required_skills" db:"required_skills"`
	LocationLat      *float64               `json:"location_lat" db:"location_lat"`
	LocationLng      *float64               `json:"location_lng" db:"location_lng"`
	LocationAddress  string                 `json:"location_address" db:"location_address"`
	StartDate        *time.Time             `json:"start_date" db:"start_date"`
	EndDate          *time.Time             `json:"end_date" db:"end_date"`
	Status           string                 `json:"status" db:"status"`
	ProjectStatus    ProjectStatus          `json:"project_status" db:"project_status"`
	CreatedByAdminID uuid.UUID              `json:"created_by_admin_id" db:"created_by_admin_id"`
	TeamLeadID       *uuid.UUID             `json:"team_lead_id" db:"team_lead_id"`
	BudgetTotal      *float64               `json:"budget_total,omitempty" db:"budget_total"`
	BudgetSpent      *float64               `json:"budget_spent,omitempty" db:"budget_spent"`
	Permissions      map[string]interface{} `json:"permissions,omitempty" db:"permissions"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
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
	query := `
		INSERT INTO projects (id, title, description, required_skills, location_lat, location_lng, 
		                     location_address, start_date, end_date, status, project_status, 
		                     created_by_admin_id, team_lead_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at, updated_at`

	project.ID = uuid.New()
	skillsJSON, err := ToJSONArray(project.RequiredSkills)
	if err != nil {
		return err
	}

	return s.db.QueryRow(query, project.ID, project.Title, project.Description, skillsJSON,
		project.LocationLat, project.LocationLng, project.LocationAddress, project.StartDate,
		project.EndDate, project.Status, project.ProjectStatus, project.CreatedByAdminID,
		project.TeamLeadID).Scan(&project.CreatedAt, &project.UpdatedAt)
}

// GetByID retrieves a project by ID
func (s *ProjectService) GetByID(id uuid.UUID) (*Project, error) {
	project := &Project{}
	var skillsJSON string
	query := `
		SELECT id, title, description, required_skills, location_lat, location_lng, 
		       location_address, start_date, end_date, status, project_status, 
		       created_by_admin_id, team_lead_id, created_at, updated_at
		FROM projects WHERE id = $1`

	err := s.db.QueryRow(query, id).Scan(&project.ID, &project.Title, &project.Description,
		&skillsJSON, &project.LocationLat, &project.LocationLng, &project.LocationAddress,
		&project.StartDate, &project.EndDate, &project.Status, &project.ProjectStatus,
		&project.CreatedByAdminID, &project.TeamLeadID, &project.CreatedAt, &project.UpdatedAt)
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
		query = `
			SELECT id, title, description, required_skills, location_lat, location_lng, 
			       location_address, start_date, end_date, status, project_status, 
			       created_by_admin_id, team_lead_id, created_at, updated_at
			FROM projects
			WHERE ($1 IS NULL OR status = $1 OR project_status = $1)
			AND required_skills && $2
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4`
		
		skillsJSON, _ := ToJSONArray(skills)
		args = []interface{}{status, skillsJSON, limit, offset}
	} else {
		// Query without skills filter
		query = `
			SELECT id, title, description, required_skills, location_lat, location_lng, 
			       location_address, start_date, end_date, status, project_status, 
			       created_by_admin_id, team_lead_id, created_at, updated_at
			FROM projects
			WHERE ($1 IS NULL OR status = $1 OR project_status = $1)
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`
		
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
			&skillsJSON, &project.LocationLat, &project.LocationLng, &project.LocationAddress,
			&project.StartDate, &project.EndDate, &project.Status, &project.ProjectStatus,
			&project.CreatedByAdminID, &project.TeamLeadID, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			log.Printf("❌ PROJECT_LIST_SCAN: Row %d scan failed: %v", rowCount, err)
			return nil, err
		}

		// Parse required skills JSON
		if err := ParseJSONArray(skillsJSON, &project.RequiredSkills); err != nil {
			log.Printf("❌ PROJECT_LIST_PARSE: Failed to parse skills for project %s: %v", project.ID, err)
			return nil, err
		}

		projects = append(projects, project)
	}

	log.Printf("✅ PROJECT_LIST_COMPLETE: Successfully processed %d projects", len(projects))
	return projects, nil
}

// ListByTeamLead retrieves projects for a specific team lead
func (s *ProjectService) ListByTeamLead(teamLeadID uuid.UUID, limit, offset int) ([]Project, error) {
	query := `
		SELECT id, title, description, required_skills, location_lat, location_lng, 
		       location_address, start_date, end_date, status, project_status, 
		       created_by_admin_id, team_lead_id, created_at, updated_at
		FROM projects
		WHERE team_lead_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(query, teamLeadID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var project Project
		var skillsJSON string
		err := rows.Scan(&project.ID, &project.Title, &project.Description,
			&skillsJSON, &project.LocationLat, &project.LocationLng, &project.LocationAddress,
			&project.StartDate, &project.EndDate, &project.Status, &project.ProjectStatus,
			&project.CreatedByAdminID, &project.TeamLeadID, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// Parse required skills JSON
		if err := ParseJSONArray(skillsJSON, &project.RequiredSkills); err != nil {
			return nil, err
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// Update updates a project
func (s *ProjectService) Update(project *Project) error {
	query := `
		UPDATE projects 
		SET title = $2, description = $3, required_skills = $4, location_lat = $5, 
		    location_lng = $6, location_address = $7, start_date = $8, end_date = $9, 
		    status = $10, project_status = $11, team_lead_id = $12, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	skillsJSON, err := ToJSONArray(project.RequiredSkills)
	if err != nil {
		return err
	}

	return s.db.QueryRow(query, project.ID, project.Title, project.Description, skillsJSON,
		project.LocationLat, project.LocationLng, project.LocationAddress, project.StartDate,
		project.EndDate, project.Status, project.ProjectStatus, project.TeamLeadID).
		Scan(&project.UpdatedAt)
}

// Delete deletes a project
func (s *ProjectService) Delete(id uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}

// GetProjectTeamMembers retrieves team members for a project
func (s *ProjectService) GetProjectTeamMembers(projectID uuid.UUID) ([]ProjectTeamMember, error) {
	query := `
		SELECT id, project_id, volunteer_id, joined_at, status, created_at, updated_at
		FROM project_team_members 
		WHERE project_id = $1
		ORDER BY joined_at`

	rows, err := s.db.Query(query, projectID)
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
	query := `
		INSERT INTO project_team_members (id, project_id, volunteer_id, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (project_id, volunteer_id) DO UPDATE SET 
		status = EXCLUDED.status, updated_at = CURRENT_TIMESTAMP`

	_, err := s.db.Exec(query, uuid.New(), projectID, volunteerID, status)
	return err
}

// UpdateTeamMemberStatus updates a team member's status
func (s *ProjectService) UpdateTeamMemberStatus(projectID, volunteerID uuid.UUID, status TeamMemberStatus) error {
	query := `
		UPDATE project_team_members 
		SET status = $3, updated_at = CURRENT_TIMESTAMP
		WHERE project_id = $1 AND volunteer_id = $2`

	result, err := s.db.Exec(query, projectID, volunteerID, status)
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
	query := `DELETE FROM project_team_members WHERE project_id = $1 AND volunteer_id = $2`
	_, err := s.db.Exec(query, projectID, volunteerID)
	return err
}

// GetProjectSignups retrieves all applications/signups for a project
func (s *ProjectService) GetProjectSignups(projectID uuid.UUID) ([]Application, error) {
	applicationService := NewApplicationService(s.db)
	return applicationService.GetApplicationsByProject(projectID)
}

// IsTeamLead checks if a user is the team lead for a project
func (s *ProjectService) IsTeamLead(projectID, userID uuid.UUID) (bool, error) {
	query := `SELECT COUNT(1) FROM projects WHERE id = $1 AND team_lead_id = $2`
	var count int
	err := s.db.QueryRow(query, projectID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// IsTeamMember checks if a user is an active team member of a project
func (s *ProjectService) IsTeamMember(projectID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT COUNT(1) 
		FROM project_team_members ptm
		JOIN volunteers v ON ptm.volunteer_id = v.id
		WHERE ptm.project_id = $1 AND v.user_id = $2 AND ptm.status = 'active'`
	var count int
	err := s.db.QueryRow(query, projectID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// AssignTeamLead assigns a team lead to a project
func (s *ProjectService) AssignTeamLead(projectID, teamLeadID uuid.UUID) error {
	query := `
		UPDATE projects 
		SET team_lead_id = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := s.db.Exec(query, projectID, teamLeadID)
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
	query := `
		SELECT id, title, description, required_skills, location_lat, location_lng, 
		       location_address, start_date, end_date, status, project_status, 
		       created_by_admin_id, team_lead_id, created_at, updated_at
		FROM projects
		WHERE project_status IN ('recruiting', 'active')
		ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var project Project
		var skillsJSON string
		err := rows.Scan(&project.ID, &project.Title, &project.Description,
			&skillsJSON, &project.LocationLat, &project.LocationLng, &project.LocationAddress,
			&project.StartDate, &project.EndDate, &project.Status, &project.ProjectStatus,
			&project.CreatedByAdminID, &project.TeamLeadID, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// Parse required skills JSON
		if err := ParseJSONArray(skillsJSON, &project.RequiredSkills); err != nil {
			return nil, err
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// GetDB returns the database connection (needed for creating other services)
func (s *ProjectService) GetDB() *sql.DB {
	return s.db
}
