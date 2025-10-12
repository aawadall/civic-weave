package models

// Query constants for UserService
const (
	userCreateQuery = `
		INSERT INTO users (id, email, password_hash, email_verified, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`

	userGetByIDQuery = `
		SELECT id, email, password_hash, email_verified, role, created_at, updated_at
		FROM users WHERE id = $1`

	userGetByEmailQuery = `
		SELECT id, email, password_hash, email_verified, role, created_at, updated_at
		FROM users WHERE email = $1`

	userUpdateQuery = `
		UPDATE users 
		SET email = $2, password_hash = $3, email_verified = $4, role = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	userVerifyEmailQuery = `UPDATE users SET email_verified = true, updated_at = CURRENT_TIMESTAMP WHERE id = $1`

	userDeleteQuery = `DELETE FROM users WHERE id = $1`

	userListAllQuery = `SELECT id, email, password_hash, email_verified, role, created_at, updated_at FROM users ORDER BY created_at DESC`
)

