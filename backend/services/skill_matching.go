package services

import (
	"database/sql"
	"fmt"
	"math"

	"civicweave/backend/models"

	"github.com/google/uuid"
)

// VolunteerSkill represents a volunteer's skill with weight for matching
type VolunteerSkill struct {
	SkillID int     `json:"skill_id"`
	Weight  float64 `json:"weight"` // 0.1 to 1.0
}

// MatchResult contains the results of skill matching calculations
type MatchResult struct {
	CosineScore       float64 `json:"cosine_score"`
	EuclideanScore    float64 `json:"euclidean_score"`
	CoverageScore     float64 `json:"coverage_score"`
	MatchedSkillIDs   []int   `json:"matched_skill_ids"`
	MissingSkillIDs   []int   `json:"missing_skill_ids"`
	MatchedSkillCount int     `json:"matched_skill_count"`
	TotalRequired     int     `json:"total_required"`
}

// SkillMatchingService handles skill matching calculations
type SkillMatchingService struct {
	db *sql.DB
}

// NewSkillMatchingService creates a new skill matching service
func NewSkillMatchingService(db *sql.DB) *SkillMatchingService {
	return &SkillMatchingService{db: db}
}

// CalculateMatch calculates match scores between volunteer skills and project requirements
func (s *SkillMatchingService) CalculateMatch(volunteerSkills []VolunteerSkill, projectSkillIDs []int) MatchResult {
	// Create map for fast lookup: skillID → weight
	vMap := make(map[int]float64)
	for _, vs := range volunteerSkills {
		vMap[vs.SkillID] = vs.Weight
	}

	// Find intersection and missing skills
	var matched []int
	var missing []int
	var vWeights []float64 // volunteer weights in project dimensions

	for _, pID := range projectSkillIDs {
		if w, exists := vMap[pID]; exists {
			matched = append(matched, pID)
			vWeights = append(vWeights, w)
		} else {
			missing = append(missing, pID)
			vWeights = append(vWeights, 0.0) // missing = weight 0
		}
	}

	normP := float64(len(projectSkillIDs))

	// 1. Cosine Similarity (restricted to intersection only)
	cosineScore := s.calculateCosineSimilarity(volunteerSkills, projectSkillIDs)

	// 2. Euclidean Distance (all project dimensions, assume initial distance = 1)
	euclideanScore := s.calculateEuclideanSimilarity(vWeights, normP)

	// 3. Weighted Coverage (simple, interpretable)
	coverageScore := s.calculateCoverageScore(vWeights, normP)

	return MatchResult{
		CosineScore:       cosineScore,
		EuclideanScore:    euclideanScore,
		CoverageScore:     coverageScore,
		MatchedSkillIDs:   matched,
		MissingSkillIDs:   missing,
		MatchedSkillCount: len(matched),
		TotalRequired:     len(projectSkillIDs),
	}
}

// calculateCosineSimilarity computes cosine similarity restricted to intersection
func (s *SkillMatchingService) calculateCosineSimilarity(volunteerSkills []VolunteerSkill, projectSkillIDs []int) float64 {
	// Find intersection skills
	var intersectionSkills []VolunteerSkill
	projectSet := make(map[int]bool)
	for _, id := range projectSkillIDs {
		projectSet[id] = true
	}

	for _, vs := range volunteerSkills {
		if projectSet[vs.SkillID] {
			intersectionSkills = append(intersectionSkills, vs)
		}
	}

	if len(intersectionSkills) == 0 {
		return 0.0
	}

	// Build vectors in intersection space
	dotProduct := 0.0
	normV := 0.0

	for _, vs := range intersectionSkills {
		dotProduct += vs.Weight // P[i] = 1 for all required skills
		normV += vs.Weight * vs.Weight
	}

	normV = math.Sqrt(normV)
	normP := math.Sqrt(float64(len(intersectionSkills))) // All P[i] = 1

	if normV > 0 && normP > 0 {
		return dotProduct / (normV * normP)
	}

	return 0.0
}

// calculateEuclideanSimilarity computes similarity based on Euclidean distance
func (s *SkillMatchingService) calculateEuclideanSimilarity(volunteerWeights []float64, normP float64) float64 {
	if normP == 0 {
		return 0.0
	}

	// Calculate distance from ideal (weight = 1.0 for all project skills)
	euclideanDist := 0.0
	for _, w := range volunteerWeights {
		euclideanDist += (w - 1.0) * (w - 1.0) // distance from ideal = 1
	}
	euclideanDist = math.Sqrt(euclideanDist)

	// Normalize to [0,1] range
	normalizedDist := euclideanDist / math.Sqrt(normP)
	if normalizedDist > 1.0 {
		normalizedDist = 1.0
	}

	return 1.0 - normalizedDist // Convert distance to similarity
}

// calculateCoverageScore computes simple weighted coverage
func (s *SkillMatchingService) calculateCoverageScore(volunteerWeights []float64, normP float64) float64 {
	if normP == 0 {
		return 0.0
	}

	weightSum := 0.0
	for _, w := range volunteerWeights {
		weightSum += w
	}

	return weightSum / normP
}

// CalculateVolunteerInitiativeMatch calculates match for a specific volunteer-initiative pair
func (s *SkillMatchingService) CalculateVolunteerInitiativeMatch(volunteerID, initiativeID uuid.UUID) (*MatchResult, error) {
	// Get volunteer skills
	volunteerSkills, err := s.getVolunteerSkills(volunteerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get volunteer skills: %w", err)
	}

	// Get initiative required skills
	projectSkills, err := s.getInitiativeSkills(initiativeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get initiative skills: %w", err)
	}

	// Convert to matching format
	var vSkills []VolunteerSkill
	for _, vs := range volunteerSkills {
		vSkills = append(vSkills, VolunteerSkill{
			SkillID: vs.SkillID,
			Weight:  vs.SkillWeight,
		})
	}

	var pSkillIDs []int
	for _, ps := range projectSkills {
		pSkillIDs = append(pSkillIDs, ps.SkillID)
	}

	// Calculate match
	result := s.CalculateMatch(vSkills, pSkillIDs)
	return &result, nil
}

// BatchCalculateMatches calculates matches for multiple volunteer-initiative pairs
func (s *SkillMatchingService) BatchCalculateMatches() error {
	// Get all active initiatives with their required skills
	initiatives, err := s.getAllActiveInitiativesWithSkills()
	if err != nil {
		return fmt.Errorf("failed to get initiatives: %w", err)
	}

	// Get all volunteers with their skills
	volunteers, err := s.getAllVolunteersWithSkills()
	if err != nil {
		return fmt.Errorf("failed to get volunteers: %w", err)
	}

	// Clear existing matches
	_, err = s.db.Exec("TRUNCATE volunteer_initiative_matches")
	if err != nil {
		return fmt.Errorf("failed to clear existing matches: %w", err)
	}

	// Calculate matches for all combinations
	for _, initiative := range initiatives {
		for _, volunteer := range volunteers {
			result := s.CalculateMatch(volunteer.Skills, initiative.RequiredSkillIDs)

			// Only store if at least 1 skill matches
			if result.MatchedSkillCount > 0 {
				err := s.storeMatch(volunteer.ID, initiative.ID, result)
				if err != nil {
					return fmt.Errorf("failed to store match: %w", err)
				}
			}
		}
	}

	return nil
}

// storeMatch stores a calculated match in the database
func (s *SkillMatchingService) storeMatch(volunteerID, initiativeID uuid.UUID, result MatchResult) error {
	query := `
		INSERT INTO volunteer_initiative_matches 
		(volunteer_id, initiative_id, match_score, jaccard_index, 
		 matched_skill_ids, matched_skill_count, calculated_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
	`

	jaccardIndex := float64(result.MatchedSkillCount) / float64(result.TotalRequired)

	_, err := s.db.Exec(query,
		volunteerID,
		initiativeID,
		result.CosineScore, // Use cosine as primary match score
		jaccardIndex,
		result.MatchedSkillIDs,
		result.MatchedSkillCount,
	)

	return err
}

// Helper types for batch processing
type InitiativeWithSkills struct {
	ID               uuid.UUID `json:"id"`
	RequiredSkillIDs []int     `json:"required_skill_ids"`
}

type VolunteerWithSkills struct {
	ID     uuid.UUID        `json:"id"`
	Skills []VolunteerSkill `json:"skills"`
}

// getAllActiveInitiativesWithSkills retrieves all active initiatives and their required skills
func (s *SkillMatchingService) getAllActiveInitiativesWithSkills() ([]InitiativeWithSkills, error) {
	query := `
		SELECT i.id, array_agg(irs.skill_id) as required_skills
		FROM initiatives i
		JOIN initiative_required_skills irs ON i.id = irs.initiative_id
		WHERE i.status = 'active'
		GROUP BY i.id
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var initiatives []InitiativeWithSkills
	for rows.Next() {
		var initiative InitiativeWithSkills
		err := rows.Scan(&initiative.ID, &initiative.RequiredSkillIDs)
		if err != nil {
			return nil, err
		}
		initiatives = append(initiatives, initiative)
	}

	return initiatives, rows.Err()
}

// getAllVolunteersWithSkills retrieves all volunteers and their skills
func (s *SkillMatchingService) getAllVolunteersWithSkills() ([]VolunteerWithSkills, error) {
	query := `
		SELECT v.id, vs.skill_id, vs.skill_weight
		FROM volunteers v
		JOIN volunteer_skills vs ON v.id = vs.volunteer_id
		ORDER BY v.id, vs.skill_id
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	volunteerMap := make(map[uuid.UUID]*VolunteerWithSkills)

	for rows.Next() {
		var volunteerID uuid.UUID
		var skillID int
		var weight float64

		err := rows.Scan(&volunteerID, &skillID, &weight)
		if err != nil {
			return nil, err
		}

		if volunteerMap[volunteerID] == nil {
			volunteerMap[volunteerID] = &VolunteerWithSkills{
				ID:     volunteerID,
				Skills: []VolunteerSkill{},
			}
		}

		volunteerMap[volunteerID].Skills = append(volunteerMap[volunteerID].Skills, VolunteerSkill{
			SkillID: skillID,
			Weight:  weight,
		})
	}

	// Convert map to slice
	var volunteers []VolunteerWithSkills
	for _, volunteer := range volunteerMap {
		volunteers = append(volunteers, *volunteer)
	}

	return volunteers, rows.Err()
}

// getVolunteerSkills retrieves skills for a specific volunteer
func (s *SkillMatchingService) getVolunteerSkills(volunteerID uuid.UUID) ([]models.VolunteerSkill, error) {
	taxonomyService := models.NewSkillTaxonomyService(s.db)
	return taxonomyService.GetVolunteerSkills(volunteerID)
}

// getInitiativeSkills retrieves required skills for a specific initiative
func (s *SkillMatchingService) getInitiativeSkills(initiativeID uuid.UUID) ([]models.InitiativeRequiredSkill, error) {
	taxonomyService := models.NewSkillTaxonomyService(s.db)
	return taxonomyService.GetInitiativeSkills(initiativeID)
}
