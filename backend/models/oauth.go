package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// OAuthAccount represents an OAuth account linked to a user
type OAuthAccount struct {
	ID             uuid.UUID `json:"id" db:"id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	Provider       string    `json:"provider" db:"provider"`
	ProviderUserID string    `json:"provider_user_id" db:"provider_user_id"`
	AccessToken    string    `json:"-" db:"access_token"`
	RefreshToken   string    `json:"-" db:"refresh_token"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// OAuthAccountService handles OAuth account operations
type OAuthAccountService struct {
	db *sql.DB
}

// NewOAuthAccountService creates a new OAuth account service
func NewOAuthAccountService(db *sql.DB) *OAuthAccountService {
	return &OAuthAccountService{db: db}
}

// Create creates a new OAuth account
func (s *OAuthAccountService) Create(account *OAuthAccount) error {
	query := `
		INSERT INTO oauth_accounts (id, user_id, provider, provider_user_id, access_token, refresh_token)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`

	account.ID = uuid.New()
	return s.db.QueryRow(query, account.ID, account.UserID, account.Provider,
		account.ProviderUserID, account.AccessToken, account.RefreshToken).
		Scan(&account.CreatedAt)
}

// GetByProviderAndUserID retrieves an OAuth account by provider and user ID
func (s *OAuthAccountService) GetByProviderAndUserID(provider, providerUserID string) (*OAuthAccount, error) {
	account := &OAuthAccount{}
	query := `
		SELECT id, user_id, provider, provider_user_id, access_token, refresh_token, created_at
		FROM oauth_accounts WHERE provider = $1 AND provider_user_id = $2`

	err := s.db.QueryRow(query, provider, providerUserID).Scan(
		&account.ID, &account.UserID, &account.Provider, &account.ProviderUserID,
		&account.AccessToken, &account.RefreshToken, &account.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return account, nil
}

// GetByUserID retrieves OAuth accounts by user ID
func (s *OAuthAccountService) GetByUserID(userID uuid.UUID) ([]*OAuthAccount, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, access_token, refresh_token, created_at
		FROM oauth_accounts WHERE user_id = $1`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*OAuthAccount
	for rows.Next() {
		account := &OAuthAccount{}
		err := rows.Scan(
			&account.ID, &account.UserID, &account.Provider, &account.ProviderUserID,
			&account.AccessToken, &account.RefreshToken, &account.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, rows.Err()
}

// Update updates an OAuth account
func (s *OAuthAccountService) Update(account *OAuthAccount) error {
	query := `
		UPDATE oauth_accounts 
		SET access_token = $2, refresh_token = $3
		WHERE id = $1`

	_, err := s.db.Exec(query, account.ID, account.AccessToken, account.RefreshToken)
	return err
}

// Delete deletes an OAuth account
func (s *OAuthAccountService) Delete(id uuid.UUID) error {
	query := `DELETE FROM oauth_accounts WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}
