package models

// Query constants for InitiativeService
const (
	initiativeCreateQuery = `
		INSERT INTO initiatives (id, title, description, required_skills, location_lat, location_lng, 
		                        location_address, start_date, end_date, status, created_by_admin_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at`

	initiativeGetByIDQuery = `
		SELECT id, title, description, required_skills, location_lat, location_lng, 
		       location_address, start_date, end_date, status, created_by_admin_id, created_at, updated_at
		FROM initiatives WHERE id = $1`

	initiativeListQuery = `
		SELECT id, title, description, required_skills, location_lat, location_lng, 
		       location_address, start_date, end_date, status, created_by_admin_id, created_at, updated_at
		FROM initiatives 
		WHERE ($1 = '' OR status = $1)
		  AND ($2::text[] IS NULL OR required_skills ?| $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	initiativeUpdateQuery = `
		UPDATE initiatives 
		SET title = $2, description = $3, required_skills = $4, location_lat = $5, location_lng = $6, 
		    location_address = $7, start_date = $8, end_date = $9, status = $10, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	initiativeDeleteQuery = `DELETE FROM initiatives WHERE id = $1`
)

