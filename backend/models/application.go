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

// GetDB returns the database connection
func (s *ApplicationService) GetDB() *sql.DB {
	return s.db
}

// Create creates a new application
func (s *ApplicationService) Create(application *Application) error {
	application.ID = uuid.New()
	return s.db.QueryRow(applicationCreateQuery, application.ID, application.VolunteerID,
		application.ProjectID, application.Status, application.AdminNotes).
		Scan(&application.AppliedAt, &application.UpdatedAt)
}

// GetByID retrieves an application by ID
func (s *ApplicationService) GetByID(id uuid.UUID) (*Application, error) {
	application := &Application{}

	err := s.db.QueryRow(applicationGetByIDQuery, id).Scan(
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

	err := s.db.QueryRow(applicationGetByVolunteerAndProjectQuery, volunteerID, projectID).Scan(
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
	rows, err := s.db.Query(applicationListQuery, volunteerID, initiativeID, status, limit, offset)
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
	return s.db.QueryRow(applicationUpdateQuery, application.ID, application.Status, application.AdminNotes).
		Scan(&application.UpdatedAt)
}

// Delete deletes an application
func (s *ApplicationService) Delete(id uuid.UUID) error {
	_, err := s.db.Exec(applicationDeleteQuery, id)
	return err
}

// GetApplicationsByProject retrieves applications for a specific project
func (s *ApplicationService) GetApplicationsByProject(projectID uuid.UUID) ([]Application, error) {
	rows, err := s.db.Query(applicationGetByProjectQuery, projectID)
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
