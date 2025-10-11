package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SkillTaxonomy represents a skill in the global taxonomy
type SkillTaxonomy struct {
	ID        int       `json:"id" db:"id"`
	SkillName string    `json:"skill_name" db:"skill_name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// VolunteerSkill represents a volunteer's skill with weight
type VolunteerSkill struct {
	VolunteerID       uuid.UUID `json:"volunteer_id" db:"volunteer_id"`
	SkillID           int       `json:"skill_id" db:"skill_id"`
	SkillWeight       float64   `json:"skill_weight" db:"skill_weight"`
	ProficiencyLevel  *string   `json:"proficiency_level" db:"proficiency_level"`
	YearsExperience   *int      `json:"years_experience" db:"years_experience"`
	LastUsedYear      *int      `json:"last_used_year" db:"last_used_year"`
	AddedAt           time.Time `json:"added_at" db:"added_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
	// Joined fields
	SkillName         string    `json:"skill_name" db:"skill_name"`
}

// InitiativeRequiredSkill represents a required skill for an initiative
type InitiativeRequiredSkill struct {
	InitiativeID uuid.UUID `json:"initiative_id" db:"initiative_id"`
	SkillID      int       `json:"skill_id" db:"skill_id"`
	AddedAt      time.Time `json:"added_at" db:"added_at"`
	// Joined fields
	SkillName    string    `json:"skill_name" db:"skill_name"`
}

// VolunteerInitiativeMatch represents pre-calculated match scores
type VolunteerInitiativeMatch struct {
	VolunteerID        uuid.UUID `json:"volunteer_id" db:"volunteer_id"`
	InitiativeID       uuid.UUID `json:"initiative_id" db:"initiative_id"`
	MatchScore         float64   `json:"match_score" db:"match_score"`
	JaccardIndex       float64   `json:"jaccard_index" db:"jaccard_index"`
	MatchedSkillIDs    []int     `json:"matched_skill_ids" db:"matched_skill_ids"`
	MatchedSkillCount  int       `json:"matched_skill_count" db:"matched_skill_count"`
	CalculatedAt       time.Time `json:"calculated_at" db:"calculated_at"`
}

// SkillTaxonomyService handles skill taxonomy operations
type SkillTaxonomyService struct {
	db *sql.DB
}

// NewSkillTaxonomyService creates a new skill taxonomy service
func NewSkillTaxonomyService(db *sql.DB) *SkillTaxonomyService {
	return &SkillTaxonomyService{db: db}
}

// GetAllSkills retrieves all skills in the taxonomy
func (s *SkillTaxonomyService) GetAllSkills() ([]SkillTaxonomy, error) {
	query := `
		SELECT id, skill_name, created_at, updated_at
		FROM skill_taxonomy
		ORDER BY skill_name
	`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query skills: %w", err)
	}
	defer rows.Close()
	
	var skills []SkillTaxonomy
	for rows.Next() {
		var skill SkillTaxonomy
		err := rows.Scan(&skill.ID, &skill.SkillName, &skill.CreatedAt, &skill.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan skill: %w", err)
		}
		skills = append(skills, skill)
	}
	
	return skills, rows.Err()
}

// AddSkill adds a new skill to the taxonomy (if it doesn't exist)
func (s *SkillTaxonomyService) AddSkill(name string) (*SkillTaxonomy, error) {
	// First, try to find existing skill (case-insensitive)
	var existing SkillTaxonomy
	err := s.db.QueryRow(`
		SELECT id, skill_name, created_at, updated_at
		FROM skill_taxonomy
		WHERE LOWER(skill_name) = LOWER($1)
	`, name).Scan(&existing.ID, &existing.SkillName, &existing.CreatedAt, &existing.UpdatedAt)
	
	if err == nil {
		return &existing, nil // Skill already exists
	}
	
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing skill: %w", err)
	}
	
	// Create new skill
	skill := &SkillTaxonomy{
		SkillName: name,
	}
	
	query := `
		INSERT INTO skill_taxonomy (skill_name)
		VALUES ($1)
		RETURNING id, created_at, updated_at
	`
	
	err = s.db.QueryRow(query, name).Scan(&skill.ID, &skill.CreatedAt, &skill.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create skill: %w", err)
	}
	
	return skill, nil
}

// GetVolunteerSkills retrieves all skills for a volunteer with weights
func (s *SkillTaxonomyService) GetVolunteerSkills(volunteerID uuid.UUID) ([]VolunteerSkill, error) {
	query := `
		SELECT vs.volunteer_id, vs.skill_id, vs.skill_weight, 
		       vs.proficiency_level, vs.years_experience, vs.last_used_year,
		       vs.added_at, vs.updated_at, st.skill_name
		FROM volunteer_skills vs
		JOIN skill_taxonomy st ON vs.skill_id = st.id
		WHERE vs.volunteer_id = $1
		ORDER BY st.skill_name
	`
	
	rows, err := s.db.Query(query, volunteerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query volunteer skills: %w", err)
	}
	defer rows.Close()
	
	var skills []VolunteerSkill
	for rows.Next() {
		var skill VolunteerSkill
		err := rows.Scan(
			&skill.VolunteerID, &skill.SkillID, &skill.SkillWeight,
			&skill.ProficiencyLevel, &skill.YearsExperience, &skill.LastUsedYear,
			&skill.AddedAt, &skill.UpdatedAt, &skill.SkillName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan volunteer skill: %w", err)
		}
		skills = append(skills, skill)
	}
	
	return skills, rows.Err()
}

// UpdateVolunteerSkills replaces all skills for a volunteer
func (s *SkillTaxonomyService) UpdateVolunteerSkills(volunteerID uuid.UUID, skillIDs []int) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Delete existing skills
	_, err = tx.Exec(`
		DELETE FROM volunteer_skills WHERE volunteer_id = $1
	`, volunteerID)
	if err != nil {
		return fmt.Errorf("failed to delete existing skills: %w", err)
	}
	
	// Insert new skills with default weight 0.5
	for _, skillID := range skillIDs {
		_, err = tx.Exec(`
			INSERT INTO volunteer_skills (volunteer_id, skill_id, skill_weight)
			VALUES ($1, $2, $3)
		`, volunteerID, skillID, 0.5)
		if err != nil {
			return fmt.Errorf("failed to insert skill %d: %w", skillID, err)
		}
	}
	
	return tx.Commit()
}

// AddVolunteerSkills adds new skills to a volunteer (without removing existing)
func (s *SkillTaxonomyService) AddVolunteerSkills(volunteerID uuid.UUID, skillIDs []int) error {
	for _, skillID := range skillIDs {
		// Use INSERT ... ON CONFLICT DO NOTHING to avoid duplicates
		_, err := s.db.Exec(`
			INSERT INTO volunteer_skills (volunteer_id, skill_id, skill_weight)
			VALUES ($1, $2, $3)
			ON CONFLICT (volunteer_id, skill_id) DO NOTHING
		`, volunteerID, skillID, 0.5)
		if err != nil {
			return fmt.Errorf("failed to add skill %d: %w", skillID, err)
		}
	}
	return nil
}

// RemoveVolunteerSkill removes a specific skill from a volunteer
func (s *SkillTaxonomyService) RemoveVolunteerSkill(volunteerID uuid.UUID, skillID int) error {
	_, err := s.db.Exec(`
		DELETE FROM volunteer_skills 
		WHERE volunteer_id = $1 AND skill_id = $2
	`, volunteerID, skillID)
	if err != nil {
		return fmt.Errorf("failed to remove volunteer skill: %w", err)
	}
	return nil
}

// GetInitiativeSkills retrieves all required skills for an initiative
func (s *SkillTaxonomyService) GetInitiativeSkills(initiativeID uuid.UUID) ([]InitiativeRequiredSkill, error) {
	query := `
		SELECT irs.initiative_id, irs.skill_id, irs.added_at, st.skill_name
		FROM initiative_required_skills irs
		JOIN skill_taxonomy st ON irs.skill_id = st.id
		WHERE irs.initiative_id = $1
		ORDER BY st.skill_name
	`
	
	rows, err := s.db.Query(query, initiativeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query initiative skills: %w", err)
	}
	defer rows.Close()
	
	var skills []InitiativeRequiredSkill
	for rows.Next() {
		var skill InitiativeRequiredSkill
		err := rows.Scan(&skill.InitiativeID, &skill.SkillID, &skill.AddedAt, &skill.SkillName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan initiative skill: %w", err)
		}
		skills = append(skills, skill)
	}
	
	return skills, rows.Err()
}

// UpdateInitiativeSkills replaces all required skills for an initiative
func (s *SkillTaxonomyService) UpdateInitiativeSkills(initiativeID uuid.UUID, skillIDs []int) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Delete existing required skills
	_, err = tx.Exec(`
		DELETE FROM initiative_required_skills WHERE initiative_id = $1
	`, initiativeID)
	if err != nil {
		return fmt.Errorf("failed to delete existing initiative skills: %w", err)
	}
	
	// Insert new required skills
	for _, skillID := range skillIDs {
		_, err = tx.Exec(`
			INSERT INTO initiative_required_skills (initiative_id, skill_id)
			VALUES ($1, $2)
		`, initiativeID, skillID)
		if err != nil {
			return fmt.Errorf("failed to insert initiative skill %d: %w", skillID, err)
		}
	}
	
	return tx.Commit()
}

// ResolveSkillNames converts skill names to IDs, adding new skills to taxonomy if needed
func (s *SkillTaxonomyService) ResolveSkillNames(skillNames []string) ([]int, error) {
	var skillIDs []int
	
	for _, name := range skillNames {
		if name == "" {
			continue // Skip empty names
		}
		
		skill, err := s.AddSkill(name)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve skill '%s': %w", name, err)
		}
		skillIDs = append(skillIDs, skill.ID)
	}
	
	return skillIDs, nil
}

// GetVolunteerProfileCompletion calculates completion percentage for a volunteer
func (s *SkillTaxonomyService) GetVolunteerProfileCompletion(volunteerID uuid.UUID) (int, error) {
	// Check if volunteer has location
	var hasLocation bool
	err := s.db.QueryRow(`
		SELECT (location_lat IS NOT NULL AND location_lng IS NOT NULL)
		FROM volunteers WHERE id = $1
	`, volunteerID).Scan(&hasLocation)
	if err != nil {
		return 0, fmt.Errorf("failed to check location: %w", err)
	}
	
	// Check if volunteer has availability set
	var hasAvailability bool
	err = s.db.QueryRow(`
		SELECT (availability IS NOT NULL AND availability != '{}')
		FROM volunteers WHERE id = $1
	`, volunteerID).Scan(&hasAvailability)
	if err != nil {
		return 0, fmt.Errorf("failed to check availability: %w", err)
	}
	
	// Check if volunteer has skills
	var skillCount int
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM volunteer_skills WHERE volunteer_id = $1
	`, volunteerID).Scan(&skillCount)
	if err != nil {
		return 0, fmt.Errorf("failed to count skills: %w", err)
	}
	hasSkills := skillCount > 0
	
	// Calculate completion percentage
	// Skills: 40%, Location: 30%, Availability: 30%
	completion := 0
	if hasSkills {
		completion += 40
	}
	if hasLocation {
		completion += 30
	}
	if hasAvailability {
		completion += 30
	}
	
	return completion, nil
}
