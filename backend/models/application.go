package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Application represents a volunteer application for a project
type Application struct {
	ID          uuid.UUID `json:"id" db:"id"`
	VolunteerID uuid.UUID `json:"volunteer_id" db:"volunteer_id"`
	ProjectID   uuid.UUID `json:"project_id" db:"project_id"`
	Status      string    `json:"status" db:"status"`
	AppliedAt   time.Time `json:"applied_at" db:"applied_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	AdminNotes  string    `json:"admin_notes" db:"admin_notes"`
}

// ApplicationService handles application operations
type ApplicationService struct {
	db *sql.DB
}

// NewApplicationService creates a new application service
func NewApplicationService(db *sql.DB) *ApplicationService {
	return &ApplicationService{db: db}
}

// Create creates a new application
func (s *ApplicationService) Create(application *Application) error {
	query := `
		INSERT INTO applications (id, volunteer_id, project_id, status, admin_notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING applied_at, updated_at`

	application.ID = uuid.New()
	return s.db.QueryRow(query, application.ID, application.VolunteerID,
		application.ProjectID, application.Status, application.AdminNotes).
		Scan(&application.AppliedAt, &application.UpdatedAt)
}

// GetByID retrieves an application by ID
func (s *ApplicationService) GetByID(id uuid.UUID) (*Application, error) {
	application := &Application{}
	query := `
		SELECT id, volunteer_id, project_id, status, applied_at, updated_at, admin_notes
		FROM applications WHERE id = $1`

	err := s.db.QueryRow(query, id).Scan(
		&application.ID, &application.VolunteerID, &application.ProjectID,
		&application.Status, &application.AppliedAt, &application.UpdatedAt, &application.AdminNotes,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return application, nil
}

// GetByProjectAndVolunteer retrieves an application by project and volunteer IDs (recommended over the deprecated initiative version).
func (s *ApplicationService) GetByProjectAndVolunteer(projectID, volunteerID uuid.UUID) (*Application, error) {
	return s.GetByVolunteerAndProject(volunteerID, projectID)
}

// GetByInitiativeAndVolunteer retrieves an application by initiative and volunteer IDs (deprecated - use GetByProjectAndVolunteer)
func (s *ApplicationService) GetByInitiativeAndVolunteer(initiativeID, volunteerID uuid.UUID) (*Application, error) {
	return s.GetByProjectAndVolunteer(initiativeID, volunteerID)
}

// GetByVolunteerAndProject retrieves an application by volunteer and project IDs
func (s *ApplicationService) GetByVolunteerAndProject(volunteerID, projectID uuid.UUID) (*Application, error) {
	application := &Application{}
	query := `
		SELECT id, volunteer_id, project_id, status, applied_at, updated_at, admin_notes
		FROM applications WHERE volunteer_id = $1 AND project_id = $2`

	err := s.db.QueryRow(query, volunteerID, projectID).Scan(
		&application.ID, &application.VolunteerID, &application.ProjectID,
		&application.Status, &application.AppliedAt, &application.UpdatedAt, &application.AdminNotes,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return application, nil
}

// List retrieves applications with filtering
func (s *ApplicationService) List(limit, offset int, volunteerID, initiativeID *uuid.UUID, status string) ([]*Application, error) {
	query := `
		SELECT id, volunteer_id, project_id, status, applied_at, updated_at, admin_notes
		FROM applications 
		WHERE ($1::uuid IS NULL OR volunteer_id = $1)
		  AND ($2::uuid IS NULL OR project_id = $2)
		  AND ($3 = '' OR status = $3)
		ORDER BY applied_at DESC
		LIMIT $4 OFFSET $5`

	rows, err := s.db.Query(query, volunteerID, initiativeID, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applications []*Application
	for rows.Next() {
		application := &Application{}
		err := rows.Scan(
			&application.ID, &application.VolunteerID, &application.ProjectID,
			&application.Status, &application.AppliedAt, &application.UpdatedAt, &application.AdminNotes,
		)
		if err != nil {
			return nil, err
		}
		applications = append(applications, application)
	}

	return applications, rows.Err()
}

// Update updates an application
func (s *ApplicationService) Update(application *Application) error {
	query := `
		UPDATE applications 
		SET status = $2, admin_notes = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	return s.db.QueryRow(query, application.ID, application.Status, application.AdminNotes).
		Scan(&application.UpdatedAt)
}

// Delete deletes an application
func (s *ApplicationService) Delete(id uuid.UUID) error {
	query := `DELETE FROM applications WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}

// GetApplicationsByProject retrieves applications for a specific project
func (s *ApplicationService) GetApplicationsByProject(projectID uuid.UUID) ([]Application, error) {
	query := `
		SELECT id, volunteer_id, project_id, status, applied_at, updated_at, admin_notes
		FROM applications 
		WHERE project_id = $1
		ORDER BY applied_at DESC`

	rows, err := s.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applications []Application
	for rows.Next() {
		var application Application
		err := rows.Scan(&application.ID, &application.VolunteerID, &application.ProjectID,
			&application.Status, &application.AppliedAt, &application.UpdatedAt, &application.AdminNotes)
		if err != nil {
			return nil, err
		}
		applications = append(applications, application)
	}

	return applications, nil
}
