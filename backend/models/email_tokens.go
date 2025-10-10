package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// EmailVerificationToken represents an email verification token
type EmailVerificationToken struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// EmailVerificationTokenService handles email verification token operations
type EmailVerificationTokenService struct {
	db *sql.DB
}

// NewEmailVerificationTokenService creates a new email verification token service
func NewEmailVerificationTokenService(db *sql.DB) *EmailVerificationTokenService {
	return &EmailVerificationTokenService{db: db}
}

// Create creates a new email verification token
func (s *EmailVerificationTokenService) Create(token *EmailVerificationToken) error {
	query := `
		INSERT INTO email_verification_tokens (id, user_id, token, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`

	token.ID = uuid.New()
	return s.db.QueryRow(query, token.ID, token.UserID, token.Token, token.ExpiresAt).
		Scan(&token.CreatedAt)
}

// GetByToken retrieves a token by its value
func (s *EmailVerificationTokenService) GetByToken(tokenStr string) (*EmailVerificationToken, error) {
	token := &EmailVerificationToken{}
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM email_verification_tokens WHERE token = $1`

	err := s.db.QueryRow(query, tokenStr).Scan(
		&token.ID, &token.UserID, &token.Token, &token.ExpiresAt, &token.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return token, nil
}

// Delete deletes a token
func (s *EmailVerificationTokenService) Delete(tokenStr string) error {
	query := `DELETE FROM email_verification_tokens WHERE token = $1`
	_, err := s.db.Exec(query, tokenStr)
	return err
}

// CleanupExpiredTokens removes expired tokens
func (s *EmailVerificationTokenService) CleanupExpiredTokens() error {
	query := `DELETE FROM email_verification_tokens WHERE expires_at < NOW()`
	_, err := s.db.Exec(query)
	return err
}

// PasswordResetTokenService handles password reset token operations
type PasswordResetTokenService struct {
	db *sql.DB
}

// NewPasswordResetTokenService creates a new password reset token service
func NewPasswordResetTokenService(db *sql.DB) *PasswordResetTokenService {
	return &PasswordResetTokenService{db: db}
}

// Create creates a new password reset token
func (s *PasswordResetTokenService) Create(token *PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`

	token.ID = uuid.New()
	return s.db.QueryRow(query, token.ID, token.UserID, token.Token, token.ExpiresAt).
		Scan(&token.CreatedAt)
}

// GetByToken retrieves a token by its value
func (s *PasswordResetTokenService) GetByToken(tokenStr string) (*PasswordResetToken, error) {
	token := &PasswordResetToken{}
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM password_reset_tokens WHERE token = $1`

	err := s.db.QueryRow(query, tokenStr).Scan(
		&token.ID, &token.UserID, &token.Token, &token.ExpiresAt, &token.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return token, nil
}

// Delete deletes a token
func (s *PasswordResetTokenService) Delete(tokenStr string) error {
	query := `DELETE FROM password_reset_tokens WHERE token = $1`
	_, err := s.db.Exec(query, tokenStr)
	return err
}

// CleanupExpiredTokens removes expired tokens
func (s *PasswordResetTokenService) CleanupExpiredTokens() error {
	query := `DELETE FROM password_reset_tokens WHERE expires_at < NOW()`
	_, err := s.db.Exec(query)
	return err
}
