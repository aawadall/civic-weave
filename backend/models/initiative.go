package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Initiative represents an initiative in the system
type Initiative struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	Title            string     `json:"title" db:"title"`
	Description      string     `json:"description" db:"description"`
	RequiredSkills   []string   `json:"required_skills" db:"required_skills"`
	LocationLat      *float64   `json:"location_lat" db:"location_lat"`
	LocationLng      *float64   `json:"location_lng" db:"location_lng"`
	LocationAddress  string     `json:"location_address" db:"location_address"`
	StartDate        *time.Time `json:"start_date" db:"start_date"`
	EndDate          *time.Time `json:"end_date" db:"end_date"`
	Status           string     `json:"status" db:"status"`
	CreatedByAdminID uuid.UUID  `json:"created_by_admin_id" db:"created_by_admin_id"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
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

	initiative.ID = uuid.New()
	return s.db.QueryRow(initiativeCreateQuery, initiative.ID, initiative.Title, initiative.Description,
		skillsJSON, initiative.LocationLat, initiative.LocationLng, initiative.LocationAddress,
		initiative.StartDate, initiative.EndDate, initiative.Status, initiative.CreatedByAdminID).
		Scan(&initiative.CreatedAt, &initiative.UpdatedAt)
}

// GetByID retrieves an initiative by ID
func (s *InitiativeService) GetByID(id uuid.UUID) (*Initiative, error) {
	initiative := &Initiative{}
	var skillsJSON []byte

	err := s.db.QueryRow(initiativeGetByIDQuery, id).Scan(
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
	rows, err := s.db.Query(initiativeListQuery, status, skills, limit, offset)
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

	return s.db.QueryRow(initiativeUpdateQuery, initiative.ID, initiative.Title, initiative.Description,
		skillsJSON, initiative.LocationLat, initiative.LocationLng, initiative.LocationAddress,
		initiative.StartDate, initiative.EndDate, initiative.Status).
		Scan(&initiative.UpdatedAt)
}

// Delete deletes an initiative
func (s *InitiativeService) Delete(id uuid.UUID) error {
	_, err := s.db.Exec(initiativeDeleteQuery, id)
	return err
}
