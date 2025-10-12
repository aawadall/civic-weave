package models

// Query constants for AdminService
const (
	adminCreateQuery = `
		INSERT INTO admins (id, user_id, name)
		VALUES ($1, $2, $3)
		RETURNING created_at`

	adminGetByIDQuery = `
		SELECT id, user_id, name, created_at
		FROM admins WHERE id = $1`

	adminGetByUserIDQuery = `
		SELECT id, user_id, name, created_at
		FROM admins WHERE user_id = $1`

	adminListQuery = `
		SELECT id, user_id, name, created_at
		FROM admins 
		ORDER BY created_at DESC`

	adminUpdateQuery = `
		UPDATE admins 
		SET name = $2
		WHERE id = $1`

	adminDeleteQuery = `DELETE FROM admins WHERE id = $1`
)
