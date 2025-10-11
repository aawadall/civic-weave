package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// RatingType represents the type of rating
type RatingType string

const (
	RatingUp      RatingType = "up"
	RatingDown    RatingType = "down"
	RatingNeutral RatingType = "neutral"
)

// VolunteerRating represents a volunteer rating
type VolunteerRating struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	VolunteerID   uuid.UUID  `json:"volunteer_id" db:"volunteer_id"`
	SkillClaimID  *uuid.UUID `json:"skill_claim_id" db:"skill_claim_id"`
	RatedByUserID uuid.UUID  `json:"rated_by_user_id" db:"rated_by_user_id"`
	ProjectID     *uuid.UUID `json:"project_id" db:"project_id"`
	Rating        RatingType `json:"rating" db:"rating"`
	Notes         string     `json:"notes" db:"notes"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// VolunteerScorecard represents aggregated rating data for a volunteer
type VolunteerScorecard struct {
	VolunteerID    uuid.UUID     `json:"volunteer_id"`
	TotalRatings   int           `json:"total_ratings"`
	UpRatings      int           `json:"up_ratings"`
	DownRatings    int           `json:"down_ratings"`
	NeutralRatings int           `json:"neutral_ratings"`
	OverallScore   float64       `json:"overall_score"` // Calculated score from -1 to 1
	Skills         []SkillRating `json:"skills"`
}

// SkillRating represents rating data for a specific skill
type SkillRating struct {
	SkillClaimID   uuid.UUID `json:"skill_claim_id"`
	SkillText      string    `json:"skill_text"`
	TotalRatings   int       `json:"total_ratings"`
	UpRatings      int       `json:"up_ratings"`
	DownRatings    int       `json:"down_ratings"`
	NeutralRatings int       `json:"neutral_ratings"`
	Score          float64   `json:"score"` // Calculated score from -1 to 1
}

// VolunteerRatingService handles volunteer rating operations
type VolunteerRatingService struct {
	db *sql.DB
}

// NewVolunteerRatingService creates a new volunteer rating service
func NewVolunteerRatingService(db *sql.DB) *VolunteerRatingService {
	return &VolunteerRatingService{db: db}
}

// CreateRating creates a new volunteer rating
func (s *VolunteerRatingService) CreateRating(rating *VolunteerRating) error {
	query := `
		INSERT INTO volunteer_ratings (id, volunteer_id, skill_claim_id, rated_by_user_id, project_id, rating, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at`

	rating.ID = uuid.New()
	return s.db.QueryRow(query, rating.ID, rating.VolunteerID, rating.SkillClaimID,
		rating.RatedByUserID, rating.ProjectID, rating.Rating, rating.Notes).
		Scan(&rating.CreatedAt)
}

// GetRatingByID retrieves a rating by ID
func (s *VolunteerRatingService) GetRatingByID(id uuid.UUID) (*VolunteerRating, error) {
	rating := &VolunteerRating{}
	query := `
		SELECT id, volunteer_id, skill_claim_id, rated_by_user_id, project_id, rating, notes, created_at
		FROM volunteer_ratings WHERE id = $1`

	err := s.db.QueryRow(query, id).Scan(&rating.ID, &rating.VolunteerID, &rating.SkillClaimID,
		&rating.RatedByUserID, &rating.ProjectID, &rating.Rating, &rating.Notes, &rating.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return rating, nil
}

// ListRatingsForVolunteer retrieves all ratings for a volunteer
func (s *VolunteerRatingService) ListRatingsForVolunteer(volunteerID uuid.UUID, limit, offset int) ([]VolunteerRating, error) {
	query := `
		SELECT id, volunteer_id, skill_claim_id, rated_by_user_id, project_id, rating, notes, created_at
		FROM volunteer_ratings 
		WHERE volunteer_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(query, volunteerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []VolunteerRating
	for rows.Next() {
		var rating VolunteerRating
		err := rows.Scan(&rating.ID, &rating.VolunteerID, &rating.SkillClaimID,
			&rating.RatedByUserID, &rating.ProjectID, &rating.Rating, &rating.Notes, &rating.CreatedAt)
		if err != nil {
			return nil, err
		}
		ratings = append(ratings, rating)
	}

	return ratings, nil
}

// ListRatingsByRater retrieves all ratings given by a specific user
func (s *VolunteerRatingService) ListRatingsByRater(raterID uuid.UUID, limit, offset int) ([]VolunteerRating, error) {
	query := `
		SELECT id, volunteer_id, skill_claim_id, rated_by_user_id, project_id, rating, notes, created_at
		FROM volunteer_ratings 
		WHERE rated_by_user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(query, raterID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []VolunteerRating
	for rows.Next() {
		var rating VolunteerRating
		err := rows.Scan(&rating.ID, &rating.VolunteerID, &rating.SkillClaimID,
			&rating.RatedByUserID, &rating.ProjectID, &rating.Rating, &rating.Notes, &rating.CreatedAt)
		if err != nil {
			return nil, err
		}
		ratings = append(ratings, rating)
	}

	return ratings, nil
}

// GetVolunteerScorecard retrieves aggregated rating data for a volunteer
func (s *VolunteerRatingService) GetVolunteerScorecard(volunteerID uuid.UUID) (*VolunteerScorecard, error) {
	scorecard := &VolunteerScorecard{
		VolunteerID: volunteerID,
		Skills:      []SkillRating{},
	}

	// Get overall rating counts
	query := `
		SELECT 
			COUNT(*) as total_ratings,
			COUNT(CASE WHEN rating = 'up' THEN 1 END) as up_ratings,
			COUNT(CASE WHEN rating = 'down' THEN 1 END) as down_ratings,
			COUNT(CASE WHEN rating = 'neutral' THEN 1 END) as neutral_ratings
		FROM volunteer_ratings 
		WHERE volunteer_id = $1`

	var totalRatings, upRatings, downRatings, neutralRatings int
	err := s.db.QueryRow(query, volunteerID).Scan(&totalRatings, &upRatings, &downRatings, &neutralRatings)
	if err != nil {
		return nil, err
	}

	scorecard.TotalRatings = totalRatings
	scorecard.UpRatings = upRatings
	scorecard.DownRatings = downRatings
	scorecard.NeutralRatings = neutralRatings

	// Calculate overall score (-1 to 1)
	if totalRatings > 0 {
		scorecard.OverallScore = float64(upRatings-downRatings) / float64(totalRatings)
	}

	// Get ratings by skill
	skillQuery := `
		SELECT 
			sc.id as skill_claim_id,
			sc.claim_text as skill_text,
			COUNT(vr.id) as total_ratings,
			COUNT(CASE WHEN vr.rating = 'up' THEN 1 END) as up_ratings,
			COUNT(CASE WHEN vr.rating = 'down' THEN 1 END) as down_ratings,
			COUNT(CASE WHEN vr.rating = 'neutral' THEN 1 END) as neutral_ratings
		FROM skill_claims sc
		LEFT JOIN volunteer_ratings vr ON sc.id = vr.skill_claim_id AND vr.volunteer_id = $1
		WHERE sc.volunteer_id = $1 AND sc.is_active = true
		GROUP BY sc.id, sc.claim_text
		HAVING COUNT(vr.id) > 0
		ORDER BY sc.claim_text`

	skillRows, err := s.db.Query(skillQuery, volunteerID)
	if err != nil {
		return nil, err
	}
	defer skillRows.Close()

	for skillRows.Next() {
		var skillRating SkillRating
		var total, up, down, neutral int
		err := skillRows.Scan(&skillRating.SkillClaimID, &skillRating.SkillText,
			&total, &up, &down, &neutral)
		if err != nil {
			return nil, err
		}

		skillRating.TotalRatings = total
		skillRating.UpRatings = up
		skillRating.DownRatings = down
		skillRating.NeutralRatings = neutral

		// Calculate skill score (-1 to 1)
		if total > 0 {
			skillRating.Score = float64(up-down) / float64(total)
		}

		scorecard.Skills = append(scorecard.Skills, skillRating)
	}

	return scorecard, nil
}

// GetRatingsByProject retrieves all ratings for a specific project
func (s *VolunteerRatingService) GetRatingsByProject(projectID uuid.UUID) ([]VolunteerRating, error) {
	query := `
		SELECT id, volunteer_id, skill_claim_id, rated_by_user_id, project_id, rating, notes, created_at
		FROM volunteer_ratings 
		WHERE project_id = $1
		ORDER BY created_at DESC`

	rows, err := s.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []VolunteerRating
	for rows.Next() {
		var rating VolunteerRating
		err := rows.Scan(&rating.ID, &rating.VolunteerID, &rating.SkillClaimID,
			&rating.RatedByUserID, &rating.ProjectID, &rating.Rating, &rating.Notes, &rating.CreatedAt)
		if err != nil {
			return nil, err
		}
		ratings = append(ratings, rating)
	}

	return ratings, nil
}

// UpdateRating updates a volunteer rating
func (s *VolunteerRatingService) UpdateRating(rating *VolunteerRating) error {
	query := `
		UPDATE volunteer_ratings 
		SET rating = $2, notes = $3
		WHERE id = $1`

	_, err := s.db.Exec(query, rating.ID, rating.Rating, rating.Notes)
	return err
}

// DeleteRating deletes a volunteer rating
func (s *VolunteerRatingService) DeleteRating(id uuid.UUID) error {
	query := `DELETE FROM volunteer_ratings WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}

// HasRated checks if a user has already rated a volunteer for a specific skill/project
func (s *VolunteerRatingService) HasRated(raterID, volunteerID uuid.UUID, skillClaimID *uuid.UUID, projectID *uuid.UUID) (bool, error) {
	query := `
		SELECT COUNT(1)
		FROM volunteer_ratings 
		WHERE rated_by_user_id = $1 AND volunteer_id = $2 
		AND ($3 IS NULL OR skill_claim_id = $3)
		AND ($4 IS NULL OR project_id = $4)`

	var count int
	err := s.db.QueryRow(query, raterID, volunteerID, skillClaimID, projectID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetTopRatedVolunteers retrieves volunteers with highest overall scores
func (s *VolunteerRatingService) GetTopRatedVolunteers(limit int) ([]VolunteerWithScore, error) {
	query := `
		SELECT 
			v.id,
			v.user_id,
			v.name,
			v.phone,
			v.location_lat,
			v.location_lng,
			v.location_address,
			v.skills,
			v.availability,
			v.consent_given,
			v.created_at,
			v.updated_at,
			COALESCE(score_data.overall_score, 0) as overall_score,
			COALESCE(score_data.total_ratings, 0) as total_ratings
		FROM volunteers v
		LEFT JOIN (
			SELECT 
				volunteer_id,
				CAST(COUNT(CASE WHEN rating = 'up' THEN 1 END) - COUNT(CASE WHEN rating = 'down' THEN 1 END) AS FLOAT) / NULLIF(COUNT(*), 0) as overall_score,
				COUNT(*) as total_ratings
			FROM volunteer_ratings
			GROUP BY volunteer_id
			HAVING COUNT(*) >= 3  -- Only include volunteers with at least 3 ratings
		) score_data ON v.id = score_data.volunteer_id
		WHERE score_data.overall_score IS NOT NULL
		ORDER BY score_data.overall_score DESC, score_data.total_ratings DESC
		LIMIT $1`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var volunteers []VolunteerWithScore
	for rows.Next() {
		var volunteer VolunteerWithScore
		err := rows.Scan(&volunteer.ID, &volunteer.UserID, &volunteer.Name, &volunteer.Phone,
			&volunteer.LocationLat, &volunteer.LocationLng, &volunteer.LocationAddress,
			&volunteer.Skills, &volunteer.Availability, &volunteer.ConsentGiven,
			&volunteer.CreatedAt, &volunteer.UpdatedAt, &volunteer.OverallScore, &volunteer.TotalRatings)
		if err != nil {
			return nil, err
		}
		volunteers = append(volunteers, volunteer)
	}

	return volunteers, nil
}

// VolunteerWithScore represents a volunteer with their rating score
type VolunteerWithScore struct {
	Volunteer
	OverallScore float64 `json:"overall_score"`
	TotalRatings int     `json:"total_ratings"`
}
