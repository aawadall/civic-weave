package services

import (
	"civicweave/backend/models"
	"database/sql"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

// VectorAggregationService handles volunteer skill vector aggregation
type VectorAggregationService struct {
	db                *sql.DB
	skillClaimService *models.SkillClaimService
}

// NewVectorAggregationService creates a new vector aggregation service
func NewVectorAggregationService(db *sql.DB, skillClaimService *models.SkillClaimService) *VectorAggregationService {
	return &VectorAggregationService{
		db:                db,
		skillClaimService: skillClaimService,
	}
}

// AggregateVolunteerVector creates or updates the aggregated skill vector for a volunteer
func (s *VectorAggregationService) AggregateVolunteerVector(volunteerID uuid.UUID) error {
	// Get all active skill claims for the volunteer
	claims, err := s.skillClaimService.GetActiveClaimsByVolunteer(volunteerID)
	if err != nil {
		return fmt.Errorf("failed to get skill claims: %w", err)
	}

	// If no claims, create zero vector
	if len(claims) == 0 {
		return s.createZeroVector(volunteerID)
	}

	// Calculate weighted average of embeddings
	aggregatedVector, err := s.calculateWeightedAverage(claims)
	if err != nil {
		return fmt.Errorf("failed to calculate weighted average: %w", err)
	}

	// Get volunteer location for geographic context
	locationPoint, err := s.getVolunteerLocationPoint(volunteerID)
	if err != nil {
		return fmt.Errorf("failed to get volunteer location: %w", err)
	}

	// Upsert the aggregated vector
	err = s.upsertVolunteerVector(volunteerID, aggregatedVector, locationPoint)
	if err != nil {
		return fmt.Errorf("failed to upsert volunteer vector: %w", err)
	}

	return nil
}

// calculateWeightedAverage calculates the weighted average of skill claim embeddings
func (s *VectorAggregationService) calculateWeightedAverage(claims []*models.SkillClaimWithWeight) (pgvector.Vector, error) {
	if len(claims) == 0 {
		return nil, fmt.Errorf("no claims provided")
	}

	// Get embedding dimension from first claim
	embeddingDim := len(claims[0].Embedding.Slice())
	if embeddingDim == 0 {
		return nil, fmt.Errorf("invalid embedding dimension")
	}

	// Initialize weighted sum and total weight
	weightedSum := make([]float64, embeddingDim)
	totalWeight := 0.0

	// Calculate weighted sum
	for _, claim := range claims {
		embedding := claim.Embedding.Slice()
		weight := claim.Weight.Weight

		if len(embedding) != embeddingDim {
			return nil, fmt.Errorf("embedding dimension mismatch")
		}

		for i, value := range embedding {
			weightedSum[i] += value * weight
		}
		totalWeight += weight
	}

	// Normalize by total weight
	if totalWeight == 0 {
		return nil, fmt.Errorf("total weight is zero")
	}

	normalizedVector := make([]float64, embeddingDim)
	for i, sum := range weightedSum {
		normalizedVector[i] = sum / totalWeight
	}

	return pgvector.NewVector(normalizedVector), nil
}

// createZeroVector creates a zero vector for volunteers with no skill claims
func (s *VectorAggregationService) createZeroVector(volunteerID uuid.UUID) error {
	// Create a zero vector with standard embedding dimension (384)
	zeroVector := make([]float64, 384) // text-embedding-3-small dimension
	for i := range zeroVector {
		zeroVector[i] = 0.0
	}

	vector := pgvector.NewVector(zeroVector)
	locationPoint, err := s.getVolunteerLocationPoint(volunteerID)
	if err != nil {
		return fmt.Errorf("failed to get volunteer location: %w", err)
	}

	return s.upsertVolunteerVector(volunteerID, vector, locationPoint)
}

// getVolunteerLocationPoint gets the volunteer's location as a WKT POINT
func (s *VectorAggregationService) getVolunteerLocationPoint(volunteerID uuid.UUID) (*string, error) {
	query := `
		SELECT location_lat, location_lng 
		FROM volunteers 
		WHERE id = $1`

	var lat, lng sql.NullFloat64
	err := s.db.QueryRow(query, volunteerID).Scan(&lat, &lng)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No location data
		}
		return nil, fmt.Errorf("failed to get volunteer location: %w", err)
	}

	// If no location data, return nil
	if !lat.Valid || !lng.Valid {
		return nil, nil
	}

	// Create WKT POINT string
	wktPoint := fmt.Sprintf("POINT(%f %f)", lng.Float64, lat.Float64)
	return &wktPoint, nil
}

// upsertVolunteerVector inserts or updates the volunteer's aggregated vector
func (s *VectorAggregationService) upsertVolunteerVector(volunteerID uuid.UUID, vector pgvector.Vector, locationPoint *string) error {
	query := `
		INSERT INTO volunteer_skill_vectors (volunteer_id, aggregated_vector, location_point, last_aggregated_at)
		VALUES ($1, $2, ST_GeomFromText($3, 4326), CURRENT_TIMESTAMP)
		ON CONFLICT (volunteer_id) 
		DO UPDATE SET 
			aggregated_vector = EXCLUDED.aggregated_vector,
			location_point = EXCLUDED.location_point,
			last_aggregated_at = EXCLUDED.last_aggregated_at,
			updated_at = CURRENT_TIMESTAMP`

	_, err := s.db.Exec(query, volunteerID, vector, locationPoint)
	if err != nil {
		return fmt.Errorf("failed to upsert volunteer vector: %w", err)
	}

	return nil
}

// UpdateVectorOnTaskCompletion updates skill weights based on task completion and re-aggregates
func (s *VectorAggregationService) UpdateVectorOnTaskCompletion(volunteerID uuid.UUID, taskID uuid.UUID, performanceScore int, relevantClaimIDs []uuid.UUID) error {
	if performanceScore < 1 || performanceScore > 5 {
		return fmt.Errorf("performance score must be between 1 and 5, got %d", performanceScore)
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update weights for relevant skill claims
	for _, claimID := range relevantClaimIDs {
		err = s.updateSkillWeightForTask(tx, claimID, performanceScore, taskID)
		if err != nil {
			return fmt.Errorf("failed to update skill weight for claim %s: %w", claimID, err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Re-aggregate the volunteer's vector
	err = s.AggregateVolunteerVector(volunteerID)
	if err != nil {
		return fmt.Errorf("failed to re-aggregate volunteer vector: %w", err)
	}

	return nil
}

// updateSkillWeightForTask updates a single skill claim's weight based on performance
func (s *VectorAggregationService) updateSkillWeightForTask(tx *sql.Tx, claimID uuid.UUID, performanceScore int, taskID uuid.UUID) error {
	// Calculate weight adjustment based on performance
	var weightAdjustment float64
	var updateReason string

	switch performanceScore {
	case 5, 4: // Excellent/Good performance
		weightAdjustment = 0.1
		updateReason = "task_completion_excellent"
	case 3: // Average performance
		weightAdjustment = 0.05
		updateReason = "task_completion_average"
	case 1, 2: // Poor performance
		weightAdjustment = -0.1
		updateReason = "task_completion_poor"
	default:
		return fmt.Errorf("invalid performance score: %d", performanceScore)
	}

	// Update weight with bounds checking
	query := `
		UPDATE skill_weights 
		SET weight = GREATEST(0.1, LEAST(1.0, weight + $1)),
		    last_task_id = $2,
		    update_reason = $3,
		    updated_at = CURRENT_TIMESTAMP
		WHERE skill_claim_id = $4`

	result, err := tx.Exec(query, weightAdjustment, taskID, updateReason, claimID)
	if err != nil {
		return fmt.Errorf("failed to update skill weight: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("skill claim not found: %s", claimID)
	}

	return nil
}

// GetVolunteerVector retrieves a volunteer's aggregated skill vector
func (s *VectorAggregationService) GetVolunteerVector(volunteerID uuid.UUID) (*models.VolunteerSkillVector, error) {
	query := `
		SELECT volunteer_id, aggregated_vector, ST_AsText(location_point) as location_point,
		       last_aggregated_at, created_at, updated_at
		FROM volunteer_skill_vectors 
		WHERE volunteer_id = $1`

	vector := &models.VolunteerSkillVector{}
	var embedding pgvector.Vector

	err := s.db.QueryRow(query, volunteerID).Scan(
		&vector.VolunteerID, &embedding, &vector.LocationPoint,
		&vector.LastAggregatedAt, &vector.CreatedAt, &vector.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No vector found
		}
		return nil, fmt.Errorf("failed to get volunteer vector: %w", err)
	}

	vector.AggregatedVector = embedding
	return vector, nil
}

// BatchAggregateVolunteerVectors aggregates vectors for multiple volunteers
func (s *VectorAggregationService) BatchAggregateVolunteerVectors(volunteerIDs []uuid.UUID) error {
	for _, volunteerID := range volunteerIDs {
		err := s.AggregateVolunteerVector(volunteerID)
		if err != nil {
			return fmt.Errorf("failed to aggregate vector for volunteer %s: %w", volunteerID, err)
		}
	}
	return nil
}

// TriggerAggregationOnClaimChange should be called when skill claims are created/updated/deleted
func (s *VectorAggregationService) TriggerAggregationOnClaimChange(volunteerID uuid.UUID) error {
	return s.AggregateVolunteerVector(volunteerID)
}

// TriggerAggregationOnWeightChange should be called when skill weights are updated
func (s *VectorAggregationService) TriggerAggregationOnWeightChange(volunteerID uuid.UUID) error {
	return s.AggregateVolunteerVector(volunteerID)
}

// CalculateVectorSimilarity calculates cosine similarity between two vectors
func (s *VectorAggregationService) CalculateVectorSimilarity(vector1, vector2 pgvector.Vector) (float64, error) {
	if vector1 == nil || vector2 == nil {
		return 0, fmt.Errorf("vectors cannot be nil")
	}

	// Use pgvector's built-in cosine distance
	distance := vector1.CosineDistance(vector2)

	// Convert distance to similarity (0-1 scale)
	similarity := 1 - distance

	// Ensure similarity is within bounds
	similarity = math.Max(0, math.Min(1, similarity))

	return similarity, nil
}
