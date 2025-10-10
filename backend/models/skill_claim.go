package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

// SkillClaim represents a volunteer's skill claim with embedding
type SkillClaim struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	VolunteerID uuid.UUID       `json:"volunteer_id" db:"volunteer_id"`
	ClaimText   string          `json:"claim_text" db:"claim_text"`
	Embedding   pgvector.Vector `json:"embedding" db:"embedding"`
	IsActive    bool            `json:"is_active" db:"is_active"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// SkillWeight represents the weight associated with a skill claim
type SkillWeight struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	SkillClaimID     uuid.UUID  `json:"skill_claim_id" db:"skill_claim_id"`
	Weight           float64    `json:"weight" db:"weight"`
	UpdatedByAdminID *uuid.UUID `json:"updated_by_admin_id" db:"updated_by_admin_id"`
	LastTaskID       *uuid.UUID `json:"last_task_id" db:"last_task_id"`
	UpdateReason     string     `json:"update_reason" db:"update_reason"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// SkillClaimWithWeight represents a skill claim with its associated weight
type SkillClaimWithWeight struct {
	SkillClaim
	Weight SkillWeight `json:"weight"`
}

// VolunteerSkillVector represents an aggregated skill vector for a volunteer
type VolunteerSkillVector struct {
	VolunteerID      uuid.UUID       `json:"volunteer_id" db:"volunteer_id"`
	AggregatedVector pgvector.Vector `json:"aggregated_vector" db:"aggregated_vector"`
	LocationPoint    *string         `json:"location_point" db:"location_point"` // WKT format
	LastAggregatedAt time.Time       `json:"last_aggregated_at" db:"last_aggregated_at"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
}

// InitiativeSkillRequirement represents skill requirements for an initiative
type InitiativeSkillRequirement struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	InitiativeID   uuid.UUID       `json:"initiative_id" db:"initiative_id"`
	RequiredVector pgvector.Vector `json:"required_vector" db:"required_vector"`
	Description    string          `json:"description" db:"description"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

// SkillClaimService handles skill claim operations
type SkillClaimService struct {
	db *sql.DB
}

// NewSkillClaimService creates a new skill claim service
func NewSkillClaimService(db *sql.DB) *SkillClaimService {
	return &SkillClaimService{db: db}
}

// CreateSkillClaim creates a new skill claim with embedding and initial weight
func (s *SkillClaimService) CreateSkillClaim(volunteerID uuid.UUID, claimText string, embedding pgvector.Vector) (*SkillClaim, error) {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create skill claim
	claim := &SkillClaim{
		ID:          uuid.New(),
		VolunteerID: volunteerID,
		ClaimText:   claimText,
		Embedding:   embedding,
		IsActive:    true,
	}

	query := `
		INSERT INTO skill_claims (id, volunteer_id, claim_text, embedding, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`

	err = tx.QueryRow(query, claim.ID, claim.VolunteerID, claim.ClaimText, claim.Embedding, claim.IsActive).
		Scan(&claim.CreatedAt, &claim.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create skill claim: %w", err)
	}

	// Create initial weight
	weight := &SkillWeight{
		ID:           uuid.New(),
		SkillClaimID: claim.ID,
		Weight:       0.5,
		UpdateReason: "initial",
	}

	weightQuery := `
		INSERT INTO skill_weights (id, skill_claim_id, weight, update_reason)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at`

	err = tx.QueryRow(weightQuery, weight.ID, weight.SkillClaimID, weight.Weight, weight.UpdateReason).
		Scan(&weight.CreatedAt, &weight.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create skill weight: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return claim, nil
}

// GetActiveClaimsByVolunteer retrieves all active skill claims for a volunteer
func (s *SkillClaimService) GetActiveClaimsByVolunteer(volunteerID uuid.UUID) ([]*SkillClaimWithWeight, error) {
	query := `
		SELECT sc.id, sc.volunteer_id, sc.claim_text, sc.embedding, sc.is_active, 
		       sc.created_at, sc.updated_at,
		       sw.id, sw.skill_claim_id, sw.weight, sw.updated_by_admin_id, 
		       sw.last_task_id, sw.update_reason, sw.created_at, sw.updated_at
		FROM skill_claims sc
		LEFT JOIN skill_weights sw ON sc.id = sw.skill_claim_id
		WHERE sc.volunteer_id = $1 AND sc.is_active = true
		ORDER BY sc.created_at DESC`

	rows, err := s.db.Query(query, volunteerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query skill claims: %w", err)
	}
	defer rows.Close()

	var claims []*SkillClaimWithWeight
	for rows.Next() {
		claim := &SkillClaimWithWeight{}
		var embedding pgvector.Vector

		err := rows.Scan(
			&claim.ID, &claim.VolunteerID, &claim.ClaimText, &embedding, &claim.IsActive,
			&claim.CreatedAt, &claim.UpdatedAt,
			&claim.Weight.ID, &claim.Weight.SkillClaimID, &claim.Weight.Weight,
			&claim.Weight.UpdatedByAdminID, &claim.Weight.LastTaskID, &claim.Weight.UpdateReason,
			&claim.Weight.CreatedAt, &claim.Weight.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan skill claim: %w", err)
		}

		claim.Embedding = embedding
		claims = append(claims, claim)
	}

	return claims, rows.Err()
}

// GetSkillClaimByID retrieves a skill claim by ID
func (s *SkillClaimService) GetSkillClaimByID(claimID uuid.UUID) (*SkillClaimWithWeight, error) {
	query := `
		SELECT sc.id, sc.volunteer_id, sc.claim_text, sc.embedding, sc.is_active, 
		       sc.created_at, sc.updated_at,
		       sw.id, sw.skill_claim_id, sw.weight, sw.updated_by_admin_id, 
		       sw.last_task_id, sw.update_reason, sw.created_at, sw.updated_at
		FROM skill_claims sc
		LEFT JOIN skill_weights sw ON sc.id = sw.skill_claim_id
		WHERE sc.id = $1`

	claim := &SkillClaimWithWeight{}
	var embedding pgvector.Vector

	err := s.db.QueryRow(query, claimID).Scan(
		&claim.ID, &claim.VolunteerID, &claim.ClaimText, &embedding, &claim.IsActive,
		&claim.CreatedAt, &claim.UpdatedAt,
		&claim.Weight.ID, &claim.Weight.SkillClaimID, &claim.Weight.Weight,
		&claim.Weight.UpdatedByAdminID, &claim.Weight.LastTaskID, &claim.Weight.UpdateReason,
		&claim.Weight.CreatedAt, &claim.Weight.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get skill claim: %w", err)
	}

	claim.Embedding = embedding
	return claim, nil
}

// UpdateSkillWeight updates the weight of a skill claim
func (s *SkillClaimService) UpdateSkillWeight(claimID uuid.UUID, newWeight float64, updatedByAdminID *uuid.UUID, updateReason string) error {
	if newWeight < 0.1 || newWeight > 1.0 {
		return fmt.Errorf("weight must be between 0.1 and 1.0, got %f", newWeight)
	}

	query := `
		UPDATE skill_weights 
		SET weight = $1, updated_by_admin_id = $2, update_reason = $3, updated_at = CURRENT_TIMESTAMP
		WHERE skill_claim_id = $4`

	result, err := s.db.Exec(query, newWeight, updatedByAdminID, updateReason, claimID)
	if err != nil {
		return fmt.Errorf("failed to update skill weight: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("skill claim not found")
	}

	return nil
}

// DeactivateSkillClaim soft deletes a skill claim by setting is_active to false
func (s *SkillClaimService) DeactivateSkillClaim(claimID uuid.UUID) error {
	query := `
		UPDATE skill_claims 
		SET is_active = false, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := s.db.Exec(query, claimID)
	if err != nil {
		return fmt.Errorf("failed to deactivate skill claim: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("skill claim not found")
	}

	return nil
}

// ListAllSkillClaims retrieves all skill claims with optional filtering (admin use)
func (s *SkillClaimService) ListAllSkillClaims(limit, offset int, activeOnly bool) ([]*SkillClaimWithWeight, error) {
	baseQuery := `
		SELECT sc.id, sc.volunteer_id, sc.claim_text, sc.embedding, sc.is_active, 
		       sc.created_at, sc.updated_at,
		       sw.id, sw.skill_claim_id, sw.weight, sw.updated_by_admin_id, 
		       sw.last_task_id, sw.update_reason, sw.created_at, sw.updated_at
		FROM skill_claims sc
		LEFT JOIN skill_weights sw ON sc.id = sw.skill_claim_id`

	if activeOnly {
		baseQuery += " WHERE sc.is_active = true"
	}

	baseQuery += " ORDER BY sc.created_at DESC LIMIT $1 OFFSET $2"

	rows, err := s.db.Query(baseQuery, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query skill claims: %w", err)
	}
	defer rows.Close()

	var claims []*SkillClaimWithWeight
	for rows.Next() {
		claim := &SkillClaimWithWeight{}
		var embedding pgvector.Vector

		err := rows.Scan(
			&claim.ID, &claim.VolunteerID, &claim.ClaimText, &embedding, &claim.IsActive,
			&claim.CreatedAt, &claim.UpdatedAt,
			&claim.Weight.ID, &claim.Weight.SkillClaimID, &claim.Weight.Weight,
			&claim.Weight.UpdatedByAdminID, &claim.Weight.LastTaskID, &claim.Weight.UpdateReason,
			&claim.Weight.CreatedAt, &claim.Weight.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan skill claim: %w", err)
		}

		claim.Embedding = embedding
		claims = append(claims, claim)
	}

	return claims, rows.Err()
}

// UpdateVolunteerSkillsVisibility updates the skills_visible flag for a volunteer
func (s *SkillClaimService) UpdateVolunteerSkillsVisibility(volunteerID uuid.UUID, visible bool) error {
	query := `
		UPDATE volunteers 
		SET skills_visible = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2`

	result, err := s.db.Exec(query, visible, volunteerID)
	if err != nil {
		return fmt.Errorf("failed to update skills visibility: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("volunteer not found")
	}

	return nil
}

// GetVolunteerSkillsVisibility gets the skills_visible flag for a volunteer
func (s *SkillClaimService) GetVolunteerSkillsVisibility(volunteerID uuid.UUID) (bool, error) {
	query := `SELECT skills_visible FROM volunteers WHERE id = $1`

	var visible bool
	err := s.db.QueryRow(query, volunteerID).Scan(&visible)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("volunteer not found")
		}
		return false, fmt.Errorf("failed to get skills visibility: %w", err)
	}

	return visible, nil
}
