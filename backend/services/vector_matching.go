package services

import (
	"civicweave/backend/models"
	"database/sql"
	"fmt"
	"math"

	"github.com/google/uuid"
)

// VectorMatchingService handles volunteer-initiative matching using skill vectors
type VectorMatchingService struct {
	db                       *sql.DB
	skillClaimService        *models.SkillClaimService
	vectorAggregationService *VectorAggregationService
}

// NewVectorMatchingService creates a new vector matching service
func NewVectorMatchingService(
	db *sql.DB,
	skillClaimService *models.SkillClaimService,
	vectorAggregationService *VectorAggregationService,
) *VectorMatchingService {
	return &VectorMatchingService{
		db:                       db,
		skillClaimService:        skillClaimService,
		vectorAggregationService: vectorAggregationService,
	}
}

// VectorMatchResult represents a match result with vector-based scoring
type VectorMatchResult struct {
	VolunteerID     string  `json:"volunteer_id"`
	InitiativeID    string  `json:"initiative_id"`
	SimilarityScore float64 `json:"similarity_score"`
	LocationScore   float64 `json:"location_score,omitempty"`
	FinalScore      float64 `json:"final_score"`
	Distance        float64 `json:"distance"` // Cosine distance
}

// GeographicMatchOptions represents options for geographic filtering
type GeographicMatchOptions struct {
	MaxDistanceKm float64
	InitiativeLat float64
	InitiativeLng float64
}

// FindTopKCandidates finds the top k volunteer candidates for an initiative using vector similarity
func (s *VectorMatchingService) FindTopKCandidates(initiativeID uuid.UUID, k int) ([]*VectorMatchResult, error) {
	if k <= 0 {
		k = 10 // Default limit
	}

	// Get initiative's required skill vector
	requiredVector, err := s.getInitiativeRequiredVector(initiativeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get initiative required vector: %w", err)
	}

	if requiredVector == nil {
		return nil, fmt.Errorf("initiative has no required skill vector")
	}

	// Find top k volunteers by cosine similarity
	query := `
		SELECT vsv.volunteer_id, vsv.aggregated_vector <=> $1 AS distance,
		       ST_AsText(vsv.location_point) as location_point
		FROM volunteer_skill_vectors vsv
		JOIN volunteers v ON vsv.volunteer_id = v.id
		WHERE v.skills_visible = true
		ORDER BY vsv.aggregated_vector <=> $1
		LIMIT $2`

	rows, err := s.db.Query(query, requiredVector, k)
	if err != nil {
		return nil, fmt.Errorf("failed to query volunteer vectors: %w", err)
	}
	defer rows.Close()

	var results []*VectorMatchResult
	for rows.Next() {
		var volunteerID uuid.UUID
		var distance float64
		var locationPoint sql.NullString

		err := rows.Scan(&volunteerID, &distance, &locationPoint)
		if err != nil {
			return nil, fmt.Errorf("failed to scan match result: %w", err)
		}

		// Convert distance to similarity score (0-1 scale)
		similarityScore := math.Max(0, 1-distance)

		result := &VectorMatchResult{
			VolunteerID:     volunteerID.String(),
			InitiativeID:    initiativeID.String(),
			SimilarityScore: similarityScore,
			FinalScore:      similarityScore, // Pure vector similarity
			Distance:        distance,
		}

		results = append(results, result)
	}

	return results, rows.Err()
}

// FindTopKWithGeoFilter finds top k candidates with geographic filtering
func (s *VectorMatchingService) FindTopKWithGeoFilter(initiativeID uuid.UUID, k int, geoOptions *GeographicMatchOptions) ([]*VectorMatchResult, error) {
	if k <= 0 {
		k = 10
	}

	// Get initiative's required skill vector
	requiredVector, err := s.getInitiativeRequiredVector(initiativeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get initiative required vector: %w", err)
	}

	if requiredVector == nil {
		return nil, fmt.Errorf("initiative has no required skill vector")
	}

	// Build geographic filter if provided
	var geoFilter string
	var args []interface{}
	argIndex := 1

	if geoOptions != nil && geoOptions.MaxDistanceKm > 0 {
		geoFilter = `
			AND ST_DWithin(
				vsv.location_point::geography,
				ST_Point($2, $3)::geography,
				$4 * 1000
			)`
		args = []interface{}{requiredVector, geoOptions.InitiativeLng, geoOptions.InitiativeLat, geoOptions.MaxDistanceKm}
		argIndex = 5
	} else {
		args = []interface{}{requiredVector}
	}

	// Query with geographic filtering and location scoring
	query := fmt.Sprintf(`
		SELECT vsv.volunteer_id, vsv.aggregated_vector <=> $1 AS distance,
		       ST_AsText(vsv.location_point) as location_point,
		       CASE 
		           WHEN vsv.location_point IS NULL THEN 0.5
		           WHEN ST_DWithin(vsv.location_point::geography, ST_Point($2, $3)::geography, 5000) THEN 1.0
		           WHEN ST_DWithin(vsv.location_point::geography, ST_Point($2, $3)::geography, 25000) THEN 0.8
		           WHEN ST_DWithin(vsv.location_point::geography, ST_Point($2, $3)::geography, 50000) THEN 0.6
		           WHEN ST_DWithin(vsv.location_point::geography, ST_Point($2, $3)::geography, 100000) THEN 0.4
		           ELSE 0.2
		       END as location_score
		FROM volunteer_skill_vectors vsv
		JOIN volunteers v ON vsv.volunteer_id = v.id
		WHERE v.skills_visible = true %s
		ORDER BY vsv.aggregated_vector <=> $1
		LIMIT $%d`, geoFilter, argIndex)

	args = append(args, k)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query volunteer vectors with geo filter: %w", err)
	}
	defer rows.Close()

	var results []*VectorMatchResult
	for rows.Next() {
		var volunteerID uuid.UUID
		var distance float64
		var locationPoint sql.NullString
		var locationScore float64

		err := rows.Scan(&volunteerID, &distance, &locationPoint, &locationScore)
		if err != nil {
			return nil, fmt.Errorf("failed to scan match result: %w", err)
		}

		// Convert distance to similarity score (0-1 scale)
		similarityScore := math.Max(0, 1-distance)

		// Calculate final score: 70% vector similarity + 30% location score
		finalScore := 0.7*similarityScore + 0.3*locationScore

		result := &VectorMatchResult{
			VolunteerID:     volunteerID.String(),
			InitiativeID:    initiativeID.String(),
			SimilarityScore: similarityScore,
			LocationScore:   locationScore,
			FinalScore:      finalScore,
			Distance:        distance,
		}

		results = append(results, result)
	}

	return results, rows.Err()
}

// getInitiativeRequiredVector retrieves the required skill vector for an initiative
func (s *VectorMatchingService) getInitiativeRequiredVector(initiativeID uuid.UUID) (interface{}, error) {
	query := `SELECT required_vector FROM initiative_skill_requirements WHERE initiative_id = $1`

	var vector interface{}
	err := s.db.QueryRow(query, initiativeID).Scan(&vector)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No required vector found
		}
		return nil, fmt.Errorf("failed to get initiative required vector: %w", err)
	}

	return vector, nil
}

// CreateInitiativeSkillRequirement creates or updates the skill requirement for an initiative
func (s *VectorMatchingService) CreateInitiativeSkillRequirement(initiativeID uuid.UUID, requiredVector interface{}, description string) error {
	query := `
		INSERT INTO initiative_skill_requirements (initiative_id, required_vector, description)
		VALUES ($1, $2, $3)
		ON CONFLICT (initiative_id) 
		DO UPDATE SET 
			required_vector = EXCLUDED.required_vector,
			description = EXCLUDED.description,
			updated_at = CURRENT_TIMESTAMP`

	_, err := s.db.Exec(query, initiativeID, requiredVector, description)
	if err != nil {
		return fmt.Errorf("failed to create/update initiative skill requirement: %w", err)
	}

	return nil
}

// CalculateVolunteerInitiativeMatch calculates the match score between a volunteer and initiative
func (s *VectorMatchingService) CalculateVolunteerInitiativeMatch(volunteerID, initiativeID uuid.UUID) (*VectorMatchResult, error) {
	// Get volunteer's aggregated vector
	volunteerVector, err := s.vectorAggregationService.GetVolunteerVector(volunteerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get volunteer vector: %w", err)
	}

	if volunteerVector == nil {
		return nil, fmt.Errorf("volunteer has no skill vector")
	}

	// Get initiative's required vector
	requiredVector, err := s.getInitiativeRequiredVector(initiativeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get initiative required vector: %w", err)
	}

	if requiredVector == nil {
		return nil, fmt.Errorf("initiative has no required skill vector")
	}

	// Calculate cosine similarity
	similarity, err := s.vectorAggregationService.CalculateVectorSimilarity(
		volunteerVector.AggregatedVector, requiredVector.(interface{}))
	if err != nil {
		return nil, fmt.Errorf("failed to calculate vector similarity: %w", err)
	}

	// Calculate distance for consistency
	distance := 1 - similarity

	result := &VectorMatchResult{
		VolunteerID:     volunteerID.String(),
		InitiativeID:    initiativeID.String(),
		SimilarityScore: similarity,
		FinalScore:      similarity, // Pure vector similarity for individual matches
		Distance:        distance,
	}

	return result, nil
}

// GetVolunteerMatchExplanations provides detailed explanations for volunteer-initiative matches
func (s *VectorMatchingService) GetVolunteerMatchExplanations(volunteerID, initiativeID uuid.UUID) (string, error) {
	match, err := s.CalculateVolunteerInitiativeMatch(volunteerID, initiativeID)
	if err != nil {
		return "", fmt.Errorf("failed to calculate match: %w", err)
	}

	var explanation string

	// Skill similarity explanation
	if match.SimilarityScore >= 0.8 {
		explanation += "Excellent skill match! Your skills align very well with this initiative. "
	} else if match.SimilarityScore >= 0.6 {
		explanation += "Good skill match. Your skills have strong alignment with this initiative. "
	} else if match.SimilarityScore >= 0.4 {
		explanation += "Moderate skill match. Some of your skills align with this initiative. "
	} else if match.SimilarityScore >= 0.2 {
		explanation += "Limited skill match. Few of your skills align with this initiative. "
	} else {
		explanation += "Poor skill match. Your current skills don't align well with this initiative. "
	}

	// Add specific score
	explanation += fmt.Sprintf("Match score: %.1f%%", match.SimilarityScore*100)

	return explanation, nil
}

// FindSimilarVolunteers finds volunteers with similar skill vectors to a given volunteer
func (s *VectorMatchingService) FindSimilarVolunteers(volunteerID uuid.UUID, k int, excludeInitiativeID *uuid.UUID) ([]*VectorMatchResult, error) {
	if k <= 0 {
		k = 5
	}

	// Get the target volunteer's vector
	targetVector, err := s.vectorAggregationService.GetVolunteerVector(volunteerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target volunteer vector: %w", err)
	}

	if targetVector == nil {
		return nil, fmt.Errorf("target volunteer has no skill vector")
	}

	// Build query to find similar volunteers
	baseQuery := `
		SELECT vsv.volunteer_id, vsv.aggregated_vector <=> $1 AS distance
		FROM volunteer_skill_vectors vsv
		JOIN volunteers v ON vsv.volunteer_id = v.id
		WHERE v.skills_visible = true AND vsv.volunteer_id != $2`

	args := []interface{}{targetVector.AggregatedVector, volunteerID}
	argIndex := 3

	// Exclude volunteers already applied to specific initiative if provided
	if excludeInitiativeID != nil {
		baseQuery += ` AND vsv.volunteer_id NOT IN (
			SELECT volunteer_id FROM applications 
			WHERE initiative_id = $` + fmt.Sprintf("%d", argIndex) + `
		)`
		args = append(args, *excludeInitiativeID)
		argIndex++
	}

	baseQuery += ` ORDER BY vsv.aggregated_vector <=> $1 LIMIT $` + fmt.Sprintf("%d", argIndex)
	args = append(args, k)

	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query similar volunteers: %w", err)
	}
	defer rows.Close()

	var results []*VectorMatchResult
	for rows.Next() {
		var similarVolunteerID uuid.UUID
		var distance float64

		err := rows.Scan(&similarVolunteerID, &distance)
		if err != nil {
			return nil, fmt.Errorf("failed to scan similar volunteer result: %w", err)
		}

		// Convert distance to similarity score
		similarityScore := math.Max(0, 1-distance)

		result := &VectorMatchResult{
			VolunteerID:     similarVolunteerID.String(),
			SimilarityScore: similarityScore,
			FinalScore:      similarityScore,
			Distance:        distance,
		}

		results = append(results, result)
	}

	return results, rows.Err()
}
