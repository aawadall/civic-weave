package models

// Query constants for ApplicationService
const (
	applicationCreateQuery = `
		INSERT INTO applications (id, volunteer_id, project_id, status, admin_notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING applied_at, updated_at`

	applicationGetByIDQuery = `
		SELECT id, volunteer_id, project_id, status, applied_at, updated_at, admin_notes
		FROM applications WHERE id = $1`

	applicationGetByVolunteerAndProjectQuery = `
		SELECT id, volunteer_id, project_id, status, applied_at, updated_at, admin_notes
		FROM applications WHERE volunteer_id = $1 AND project_id = $2`

	applicationListQuery = `
		SELECT id, volunteer_id, project_id, status, applied_at, updated_at, admin_notes
		FROM applications 
		WHERE ($1::uuid IS NULL OR volunteer_id = $1)
		  AND ($2::uuid IS NULL OR project_id = $2)
		  AND ($3 = '' OR status = $3)
		ORDER BY applied_at DESC
		LIMIT $4 OFFSET $5`

	applicationUpdateQuery = `
		UPDATE applications 
		SET status = $2, admin_notes = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	applicationDeleteQuery = `DELETE FROM applications WHERE id = $1`

	applicationGetByProjectQuery = `
		SELECT id, volunteer_id, project_id, status, applied_at, updated_at, admin_notes
		FROM applications 
		WHERE project_id = $1
		ORDER BY applied_at DESC`
)
