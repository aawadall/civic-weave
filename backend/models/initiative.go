package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Initiative represents an initiative in the system
type Initiative struct {
	ID                uuid.UUID       `json:"id" db:"id"`
	Title             string          `json:"title" db:"title"`
	Description       string          `json:"description" db:"description"`
	RequiredSkills    []string        `json:"required_skills" db:"required_skills"`
	LocationLat       *float64        `json:"location_lat" db:"location_lat"`
	LocationLng       *float64        `json:"location_lng" db:"location_lng"`
	LocationAddress   string          `json:"location_address" db:"location_address"`
	StartDate         *time.Time      `json:"start_date" db:"start_date"`
	EndDate           *time.Time      `json:"end_date" db:"end_date"`
	Status            string          `json:"status" db:"status"`
	CreatedByAdminID  uuid.UUID       `json:"created_by_admin_id" db:"created_by_admin_id"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
}

// InitiativeService handles initiative operations
type InitiativeService struct {
	db *sql.DB
}

// NewInitiativeService creates a new initiative service
func NewInitiativeService(db *sql.DB) *InitiativeService {
	return &InitiativeService{db: db}
}

// Create creates a new initiative
func (s *InitiativeService) Create(initiative *Initiative) error {
	skillsJSON, _ := json.Marshal(initiative.RequiredSkills)
	
	query := `
		INSERT INTO initiatives (id, title, description, required_skills, location_lat, location_lng, 
		                        location_address, start_date, end_date, status, created_by_admin_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at`

	initiative.ID = uuid.New()
	return s.db.QueryRow(query, initiative.ID, initiative.Title, initiative.Description,
		skillsJSON, initiative.LocationLat, initiative.LocationLng, initiative.LocationAddress,
		initiative.StartDate, initiative.EndDate, initiative.Status, initiative.CreatedByAdminID).
		Scan(&initiative.CreatedAt, &initiative.UpdatedAt)
}

// GetByID retrieves an initiative by ID
func (s *InitiativeService) GetByID(id uuid.UUID) (*Initiative, error) {
	initiative := &Initiative{}
	var skillsJSON []byte
	
	query := `
		SELECT id, title, description, required_skills, location_lat, location_lng, 
		       location_address, start_date, end_date, status, created_by_admin_id, created_at, updated_at
		FROM initiatives WHERE id = $1`

	err := s.db.QueryRow(query, id).Scan(
		&initiative.ID, &initiative.Title, &initiative.Description, &skillsJSON,
		&initiative.LocationLat, &initiative.LocationLng, &initiative.LocationAddress,
		&initiative.StartDate, &initiative.EndDate, &initiative.Status,
		&initiative.CreatedByAdminID, &initiative.CreatedAt, &initiative.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse skills JSON
	if err := json.Unmarshal(skillsJSON, &initiative.RequiredSkills); err != nil {
		initiative.RequiredSkills = []string{}
	}

	return initiative, nil
}

// List retrieves initiatives with filtering
func (s *InitiativeService) List(limit, offset int, status string, skills []string) ([]*Initiative, error) {
	query := `
		SELECT id, title, description, required_skills, location_lat, location_lng, 
		       location_address, start_date, end_date, status, created_by_admin_id, created_at, updated_at
		FROM initiatives 
		WHERE ($1 = '' OR status = $1)
		  AND ($2::text[] IS NULL OR required_skills ?| $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := s.db.Query(query, status, skills, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var initiatives []*Initiative
	for rows.Next() {
		initiative := &Initiative{}
		var skillsJSON []byte
		
		err := rows.Scan(
			&initiative.ID, &initiative.Title, &initiative.Description, &skillsJSON,
			&initiative.LocationLat, &initiative.LocationLng, &initiative.LocationAddress,
			&initiative.StartDate, &initiative.EndDate, &initiative.Status,
			&initiative.CreatedByAdminID, &initiative.CreatedAt, &initiative.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse skills JSON
		if err := json.Unmarshal(skillsJSON, &initiative.RequiredSkills); err != nil {
			initiative.RequiredSkills = []string{}
		}

		initiatives = append(initiatives, initiative)
	}

	return initiatives, rows.Err()
}

// Update updates an initiative
func (s *InitiativeService) Update(initiative *Initiative) error {
	skillsJSON, _ := json.Marshal(initiative.RequiredSkills)
	
	query := `
		UPDATE initiatives 
		SET title = $2, description = $3, required_skills = $4, location_lat = $5, location_lng = $6, 
		    location_address = $7, start_date = $8, end_date = $9, status = $10, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	return s.db.QueryRow(query, initiative.ID, initiative.Title, initiative.Description,
		skillsJSON, initiative.LocationLat, initiative.LocationLng, initiative.LocationAddress,
		initiative.StartDate, initiative.EndDate, initiative.Status).
		Scan(&initiative.UpdatedAt)
}

// Delete deletes an initiative
func (s *InitiativeService) Delete(id uuid.UUID) error {
	query := `DELETE FROM initiatives WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}
