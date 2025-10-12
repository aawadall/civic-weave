package models

// Query constants for VolunteerService
const (
	volunteerCreateQuery = `
		INSERT INTO volunteers (id, user_id, name, phone, location_lat, location_lng, 
		                       location_address, skills, availability, skills_visible, consent_given)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at`

	volunteerGetByIDQuery = `
		SELECT id, user_id, name, phone, location_lat, location_lng, 
		       location_address, skills, availability, skills_visible, consent_given, created_at, updated_at
		FROM volunteers WHERE id = $1`

	volunteerGetByUserIDQuery = `
		SELECT id, user_id, name, phone, location_lat, location_lng, 
		       location_address, skills, availability, skills_visible, consent_given, created_at, updated_at
		FROM volunteers WHERE user_id = $1`

	volunteerListQuery = `
		SELECT id, user_id, name, phone, location_lat, location_lng, 
		       location_address, skills, availability, skills_visible, consent_given, created_at, updated_at
		FROM volunteers 
		WHERE ($1::text[] IS NULL OR skills ?| $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	volunteerUpdateQuery = `
		UPDATE volunteers 
		SET name = $2, phone = $3, location_lat = $4, location_lng = $5, 
		    location_address = $6, skills = $7, availability = $8, skills_visible = $9, consent_given = $10, 
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	volunteerDeleteQuery = `DELETE FROM volunteers WHERE id = $1`
)
