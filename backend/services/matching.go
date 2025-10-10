package services

import (
	"civicweave/backend/models"
	"math"
	"sort"

	"github.com/google/uuid"
)

// MatchingService handles volunteer-initiative matching
type MatchingService struct {
	volunteerService  *models.VolunteerService
	initiativeService *models.InitiativeService
}

// NewMatchingService creates a new matching service
func NewMatchingService(volunteerService *models.VolunteerService, initiativeService *models.InitiativeService) *MatchingService {
	return &MatchingService{
		volunteerService:  volunteerService,
		initiativeService: initiativeService,
	}
}

// Match represents a volunteer-initiative match with score
type Match struct {
	VolunteerID  string  `json:"volunteer_id"`
	InitiativeID string  `json:"initiative_id"`
	Score        float64 `json:"score"`
	Reason       string  `json:"reason"`
	SkillMatch   int     `json:"skill_match"`
	LocationDist float64 `json:"location_distance"`
}

// MatchResult represents the result of matching
type MatchResult struct {
	VolunteerID   string  `json:"volunteer_id"`
	InitiativeID  string  `json:"initiative_id"`
	Matches       []Match `json:"matches"`
	TotalScore    float64 `json:"total_score"`
	SkillScore    float64 `json:"skill_score"`
	LocationScore float64 `json:"location_score"`
}

// GetMatchesForVolunteer finds the best initiative matches for a volunteer
func (s *MatchingService) GetMatchesForVolunteer(volunteerID string, limit int) ([]MatchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	// Parse volunteer ID
	volunteerUUID, err := uuid.Parse(volunteerID)
	if err != nil {
		return nil, err
	}

	// Get volunteer
	volunteer, err := s.volunteerService.GetByID(volunteerUUID)
	if err != nil || volunteer == nil {
		return nil, err
	}

	// Get active initiatives
	initiatives, err := s.initiativeService.List(100, 0, "active", []string{})
	if err != nil {
		return nil, err
	}

	var results []MatchResult

	for _, initiative := range initiatives {
		score, skillScore, locationScore := s.calculateMatchScore(volunteer, initiative)

		if score > 0 {
			results = append(results, MatchResult{
				VolunteerID:   volunteerID,
				InitiativeID:  initiative.ID.String(),
				TotalScore:    score,
				SkillScore:    skillScore,
				LocationScore: locationScore,
			})
		}
	}

	// Sort by total score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalScore > results[j].TotalScore
	})

	// Return top matches
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// GetMatchesForInitiative finds the best volunteer matches for an initiative
func (s *MatchingService) GetMatchesForInitiative(initiativeID string, limit int) ([]MatchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	// Parse initiative ID
	initiativeUUID, err := uuid.Parse(initiativeID)
	if err != nil {
		return nil, err
	}

	// Get initiative
	initiative, err := s.initiativeService.GetByID(initiativeUUID)
	if err != nil || initiative == nil {
		return nil, err
	}

	// Get all volunteers
	volunteers, err := s.volunteerService.List(1000, 0, []string{}, "")
	if err != nil {
		return nil, err
	}

	var results []MatchResult

	for _, volunteer := range volunteers {
		score, skillScore, locationScore := s.calculateMatchScore(volunteer, initiative)

		if score > 0 {
			results = append(results, MatchResult{
				VolunteerID:   volunteer.ID.String(),
				InitiativeID:  initiativeID,
				TotalScore:    score,
				SkillScore:    skillScore,
				LocationScore: locationScore,
			})
		}
	}

	// Sort by total score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalScore > results[j].TotalScore
	})

	// Return top matches
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// calculateMatchScore calculates the match score between a volunteer and initiative
func (s *MatchingService) calculateMatchScore(volunteer *models.Volunteer, initiative *models.Initiative) (totalScore, skillScore, locationScore float64) {
	// Skill matching (60% weight)
	skillScore = s.calculateSkillScore(volunteer.Skills, initiative.RequiredSkills)

	// Location matching (40% weight)
	locationScore = s.calculateLocationScore(volunteer.LocationLat, volunteer.LocationLng, initiative.LocationLat, initiative.LocationLng)

	// Weighted total score
	totalScore = (skillScore * 0.6) + (locationScore * 0.4)

	return totalScore, skillScore, locationScore
}

// calculateSkillScore calculates skill overlap score (0-100)
func (s *MatchingService) calculateSkillScore(volunteerSkills, requiredSkills []string) float64 {
	if len(requiredSkills) == 0 {
		return 50.0 // Neutral score if no skills required
	}

	if len(volunteerSkills) == 0 {
		return 0.0 // No match if volunteer has no skills
	}

	// Convert to maps for easier comparison
	volunteerSkillMap := make(map[string]bool)
	for _, skill := range volunteerSkills {
		volunteerSkillMap[skill] = true
	}

	// Count matching skills
	matchedSkills := 0
	for _, skill := range requiredSkills {
		if volunteerSkillMap[skill] {
			matchedSkills++
		}
	}

	// Calculate percentage
	skillMatchPercent := float64(matchedSkills) / float64(len(requiredSkills))

	// Scale to 0-100 with bonus for having extra skills
	baseScore := skillMatchPercent * 80.0 // Base score up to 80

	// Bonus for having more skills than required (up to 20 points)
	if len(volunteerSkills) > len(requiredSkills) {
		bonus := math.Min(20.0, float64(len(volunteerSkills)-len(requiredSkills))*5.0)
		baseScore += bonus
	}

	return math.Min(100.0, baseScore)
}

// calculateLocationScore calculates location proximity score (0-100)
func (s *MatchingService) calculateLocationScore(volLat, volLng, initLat, initLng *float64) float64 {
	// If no location data, return neutral score
	if volLat == nil || volLng == nil || initLat == nil || initLng == nil {
		return 50.0
	}

	// Calculate distance using Haversine formula
	distance := s.calculateDistance(*volLat, *volLng, *initLat, *initLng)

	// Convert distance to score (0-100)
	// 0km = 100 points, 50km = 50 points, 100km+ = 0 points
	if distance <= 5.0 {
		return 100.0 // Very close
	} else if distance <= 25.0 {
		return 90.0 - (distance-5.0)*2.0 // Gradual decrease
	} else if distance <= 50.0 {
		return 70.0 - (distance-25.0)*1.4 // Faster decrease
	} else if distance <= 100.0 {
		return 35.0 - (distance-50.0)*0.7 // Slower decrease
	} else {
		return 0.0 // Too far
	}
}

// calculateDistance calculates distance between two points using Haversine formula
func (s *MatchingService) calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371.0 // Earth's radius in kilometers

	// Convert to radians
	lat1Rad := lat1 * math.Pi / 180
	lng1Rad := lng1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lng2Rad := lng2 * math.Pi / 180

	// Calculate differences
	dLat := lat2Rad - lat1Rad
	dLng := lng2Rad - lng1Rad

	// Haversine formula
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// CalculateMatchScore calculates the match score between a volunteer and initiative (public method)
func (s *MatchingService) CalculateMatchScore(volunteer *models.Volunteer, initiative *models.Initiative) (totalScore, skillScore, locationScore float64) {
	return s.calculateMatchScore(volunteer, initiative)
}

// GetMatchingExplanation provides human-readable explanation of match score
func (s *MatchingService) GetMatchingExplanation(volunteer *models.Volunteer, initiative *models.Initiative) string {
	skillScore := s.calculateSkillScore(volunteer.Skills, initiative.RequiredSkills)
	locationScore := s.calculateLocationScore(volunteer.LocationLat, volunteer.LocationLng, initiative.LocationLat, initiative.LocationLng)

	var explanation string

	// Skill explanation
	if skillScore >= 80 {
		explanation += "Excellent skill match! "
	} else if skillScore >= 60 {
		explanation += "Good skill match. "
	} else if skillScore >= 40 {
		explanation += "Partial skill match. "
	} else {
		explanation += "Limited skill match. "
	}

	// Location explanation
	if locationScore >= 80 {
		explanation += "Very close location."
	} else if locationScore >= 60 {
		explanation += "Nearby location."
	} else if locationScore >= 40 {
		explanation += "Moderate distance."
	} else {
		explanation += "Far location."
	}

	return explanation
}
