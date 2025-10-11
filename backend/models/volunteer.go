package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Volunteer represents a volunteer in the system
type Volunteer struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	UserID          uuid.UUID       `json:"user_id" db:"user_id"`
	Name            string          `json:"name" db:"name"`
	Phone           string          `json:"phone" db:"phone"`
	LocationLat     *float64        `json:"location_lat" db:"location_lat"`
	LocationLng     *float64        `json:"location_lng" db:"location_lng"`
	LocationAddress string          `json:"location_address" db:"location_address"`
	Skills          []string        `json:"skills" db:"skills"`
	Availability    json.RawMessage `json:"availability" db:"availability"`
	SkillsVisible   bool            `json:"skills_visible" db:"skills_visible"`
	ConsentGiven    bool            `json:"consent_given" db:"consent_given"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// VolunteerService handles volunteer operations
type VolunteerService struct {
	db *sql.DB
}

// NewVolunteerService creates a new volunteer service
func NewVolunteerService(db *sql.DB) *VolunteerService {
	return &VolunteerService{db: db}
}

// GetDB returns the database connection
func (s *VolunteerService) GetDB() *sql.DB {
	return s.db
}

// Create creates a new volunteer
func (s *VolunteerService) Create(volunteer *Volunteer) error {
	skillsJSON, _ := json.Marshal(volunteer.Skills)

	query := `
		INSERT INTO volunteers (id, user_id, name, phone, location_lat, location_lng, 
		                       location_address, skills, availability, skills_visible, consent_given)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at`

	volunteer.ID = uuid.New()
	return s.db.QueryRow(query, volunteer.ID, volunteer.UserID, volunteer.Name,
		volunteer.Phone, volunteer.LocationLat, volunteer.LocationLng,
		volunteer.LocationAddress, skillsJSON, volunteer.Availability, volunteer.SkillsVisible, volunteer.ConsentGiven).
		Scan(&volunteer.CreatedAt, &volunteer.UpdatedAt)
}

// GetByID retrieves a volunteer by ID
func (s *VolunteerService) GetByID(id uuid.UUID) (*Volunteer, error) {
	volunteer := &Volunteer{}
	var skillsJSON []byte

	query := `
		SELECT id, user_id, name, phone, location_lat, location_lng, 
		       location_address, skills, availability, skills_visible, consent_given, created_at, updated_at
		FROM volunteers WHERE id = $1`

	err := s.db.QueryRow(query, id).Scan(
		&volunteer.ID, &volunteer.UserID, &volunteer.Name, &volunteer.Phone,
		&volunteer.LocationLat, &volunteer.LocationLng, &volunteer.LocationAddress,
		&skillsJSON, &volunteer.Availability, &volunteer.SkillsVisible, &volunteer.ConsentGiven,
		&volunteer.CreatedAt, &volunteer.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse skills JSON
	if err := json.Unmarshal(skillsJSON, &volunteer.Skills); err != nil {
		volunteer.Skills = []string{}
	}

	return volunteer, nil
}

// GetByUserID retrieves a volunteer by user ID
func (s *VolunteerService) GetByUserID(userID uuid.UUID) (*Volunteer, error) {
	volunteer := &Volunteer{}
	var skillsJSON []byte

	query := `
		SELECT id, user_id, name, phone, location_lat, location_lng, 
		       location_address, skills, availability, skills_visible, consent_given, created_at, updated_at
		FROM volunteers WHERE user_id = $1`

	err := s.db.QueryRow(query, userID).Scan(
		&volunteer.ID, &volunteer.UserID, &volunteer.Name, &volunteer.Phone,
		&volunteer.LocationLat, &volunteer.LocationLng, &volunteer.LocationAddress,
		&skillsJSON, &volunteer.Availability, &volunteer.SkillsVisible, &volunteer.ConsentGiven,
		&volunteer.CreatedAt, &volunteer.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse skills JSON
	if err := json.Unmarshal(skillsJSON, &volunteer.Skills); err != nil {
		volunteer.Skills = []string{}
	}

	return volunteer, nil
}

// List retrieves volunteers with filtering
func (s *VolunteerService) List(limit, offset int, skills []string, location string) ([]*Volunteer, error) {
	query := `
		SELECT id, user_id, name, phone, location_lat, location_lng, 
		       location_address, skills, availability, skills_visible, consent_given, created_at, updated_at
		FROM volunteers 
		WHERE ($1::text[] IS NULL OR skills ?| $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(query, skills, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var volunteers []*Volunteer
	for rows.Next() {
		volunteer := &Volunteer{}
		var skillsJSON []byte

		err := rows.Scan(
			&volunteer.ID, &volunteer.UserID, &volunteer.Name, &volunteer.Phone,
			&volunteer.LocationLat, &volunteer.LocationLng, &volunteer.LocationAddress,
			&skillsJSON, &volunteer.Availability, &volunteer.SkillsVisible, &volunteer.ConsentGiven,
			&volunteer.CreatedAt, &volunteer.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse skills JSON
		if err := json.Unmarshal(skillsJSON, &volunteer.Skills); err != nil {
			volunteer.Skills = []string{}
		}

		volunteers = append(volunteers, volunteer)
	}

	return volunteers, rows.Err()
}

// GetVolunteerScorecard retrieves the scorecard for a volunteer
func (s *VolunteerService) GetVolunteerScorecard(volunteerID uuid.UUID) (*VolunteerScorecard, error) {
	ratingService := NewVolunteerRatingService(s.db)
	return ratingService.GetVolunteerScorecard(volunteerID)
}

// GetVolunteerRatings retrieves all ratings for a volunteer
func (s *VolunteerService) GetVolunteerRatings(volunteerID uuid.UUID, limit, offset int) ([]VolunteerRating, error) {
	ratingService := NewVolunteerRatingService(s.db)
	return ratingService.ListRatingsForVolunteer(volunteerID, limit, offset)
}

// GetVolunteerWithScore retrieves a volunteer with their rating score
func (s *VolunteerService) GetVolunteerWithScore(volunteerID uuid.UUID) (*VolunteerWithScore, error) {
	volunteer, err := s.GetByID(volunteerID)
	if err != nil {
		return nil, err
	}
	if volunteer == nil {
		return nil, nil
	}

	ratingService := NewVolunteerRatingService(s.db)
	scorecard, err := ratingService.GetVolunteerScorecard(volunteerID)
	if err != nil {
		return nil, err
	}

	return &VolunteerWithScore{
		Volunteer:    *volunteer,
		OverallScore: scorecard.OverallScore,
		TotalRatings: scorecard.TotalRatings,
	}, nil
}

// Update updates a volunteer
func (s *VolunteerService) Update(volunteer *Volunteer) error {
	skillsJSON, _ := json.Marshal(volunteer.Skills)

	query := `
		UPDATE volunteers 
		SET name = $2, phone = $3, location_lat = $4, location_lng = $5, 
		    location_address = $6, skills = $7, availability = $8, skills_visible = $9, consent_given = $10, 
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	return s.db.QueryRow(query, volunteer.ID, volunteer.Name, volunteer.Phone,
		volunteer.LocationLat, volunteer.LocationLng, volunteer.LocationAddress,
		skillsJSON, volunteer.Availability, volunteer.SkillsVisible, volunteer.ConsentGiven).
		Scan(&volunteer.UpdatedAt)
}

// Delete deletes a volunteer
func (s *VolunteerService) Delete(id uuid.UUID) error {
	query := `DELETE FROM volunteers WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}
