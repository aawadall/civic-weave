package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

// Validation errors
var (
	ErrInvalidEmail    = fmt.Errorf("invalid email format")
	ErrInvalidPassword = fmt.Errorf("password must be at least 8 characters")
	ErrInvalidRole     = fmt.Errorf("invalid role")
	ErrInvalidUUID     = fmt.Errorf("invalid UUID format")
	ErrEmptyField      = fmt.Errorf("required field cannot be empty")
	ErrInvalidSkills   = fmt.Errorf("invalid skills format")
	ErrInvalidStatus   = fmt.Errorf("invalid status")
)

// Email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateUser validates user input
func ValidateUser(user *User) error {
	if err := ValidateEmail(user.Email); err != nil {
		return err
	}

	if user.PasswordHash != "" {
		if err := ValidatePassword(user.PasswordHash); err != nil {
			return err
		}
	}

	if err := ValidateRole(user.Role); err != nil {
		return err
	}

	return nil
}

// ValidateVolunteer validates volunteer input
func ValidateVolunteer(volunteer *Volunteer) error {
	if strings.TrimSpace(volunteer.Name) == "" {
		return fmt.Errorf("name: %w", ErrEmptyField)
	}

	if err := ValidateSkills(volunteer.Skills); err != nil {
		return err
	}

	// Validate phone if provided
	if volunteer.Phone != "" {
		if err := ValidatePhone(volunteer.Phone); err != nil {
			return err
		}
	}

	return nil
}

// ValidateInitiative validates initiative input
func ValidateInitiative(initiative *Initiative) error {
	if strings.TrimSpace(initiative.Title) == "" {
		return fmt.Errorf("title: %w", ErrEmptyField)
	}

	if err := ValidateSkills(initiative.RequiredSkills); err != nil {
		return err
	}

	if err := ValidateStatus(initiative.Status); err != nil {
		return err
	}

	return nil
}

// ValidateApplication validates application input
func ValidateApplication(application *Application) error {
	if err := ValidateUUID(application.VolunteerID); err != nil {
		return fmt.Errorf("volunteer_id: %w", err)
	}

	if err := ValidateUUID(application.ProjectID); err != nil {
		return fmt.Errorf("project_id: %w", err)
	}

	if err := ValidateStatus(application.Status); err != nil {
		return err
	}

	return nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("email: %w", ErrEmptyField)
	}

	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	// Additional checks
	if len(email) > 255 {
		return fmt.Errorf("email: too long (max 255 characters)")
	}

	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrInvalidPassword
	}

	// Check for at least one uppercase, lowercase, and number
	var hasUpper, hasLower, hasNumber bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber {
		return fmt.Errorf("password must contain at least one uppercase letter, one lowercase letter, and one number")
	}

	return nil
}

// ValidateRole validates user role
func ValidateRole(role string) error {
	validRoles := []string{"admin", "volunteer"}
	for _, validRole := range validRoles {
		if role == validRole {
			return nil
		}
	}
	return ErrInvalidRole
}

// ValidateStatus validates initiative/application status
func ValidateStatus(status string) error {
	validStatuses := []string{"draft", "active", "closed", "pending", "accepted", "rejected"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}
	return ErrInvalidStatus
}

// ValidateSkills validates skills array
func ValidateSkills(skills []string) error {
	if len(skills) == 0 {
		return nil // Empty skills is allowed
	}

	for i, skill := range skills {
		if strings.TrimSpace(skill) == "" {
			return fmt.Errorf("skill %d: %w", i, ErrEmptyField)
		}

		if len(skill) > 100 {
			return fmt.Errorf("skill %d: too long (max 100 characters)", i)
		}

		// Remove duplicates by checking against previous skills
		for j := 0; j < i; j++ {
			if strings.EqualFold(skill, skills[j]) {
				return fmt.Errorf("duplicate skill: %s", skill)
			}
		}
	}

	return nil
}

// ValidatePhone validates phone number format
func ValidatePhone(phone string) error {
	// Remove all non-digit characters for validation
	digits := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	if len(digits) < 10 || len(digits) > 15 {
		return fmt.Errorf("invalid phone number format")
	}

	return nil
}

// ValidateUUID validates UUID format
func ValidateUUID(id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidUUID
	}
	return nil
}

// SanitizeString removes potentially dangerous characters
func SanitizeString(input string) string {
	// Remove null bytes and control characters
	input = strings.ReplaceAll(input, "\x00", "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	return input
}

// SanitizeSkills sanitizes skills array
func SanitizeSkills(skills []string) []string {
	var sanitized []string
	for _, skill := range skills {
		sanitizedSkill := SanitizeString(skill)
		if sanitizedSkill != "" {
			sanitized = append(sanitized, sanitizedSkill)
		}
	}
	return sanitized
}

// ToJSONArray converts a string slice to JSON array string
func ToJSONArray(slice []string) (string, error) {
	if slice == nil {
		return "[]", nil
	}
	jsonData, err := json.Marshal(slice)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// ParseJSONArray parses a JSON array string into a string slice
func ParseJSONArray(jsonStr string, target *[]string) error {
	if jsonStr == "" {
		*target = []string{}
		return nil
	}
	return json.Unmarshal([]byte(jsonStr), target)
}

// ParseJSONMap parses a JSONB byte array into a map[string]interface{}
func ParseJSONMap(jsonBytes []byte, target *map[string]interface{}) error {
	if len(jsonBytes) == 0 {
		return nil
	}
	return json.Unmarshal(jsonBytes, target)
}
