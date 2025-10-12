package models

// Query constants for OAuthAccountService
const (
	oauthCreateQuery = `
		INSERT INTO oauth_accounts (id, user_id, provider, provider_user_id, access_token, refresh_token)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`

	oauthGetByProviderAndUserIDQuery = `
		SELECT id, user_id, provider, provider_user_id, access_token, refresh_token, created_at
		FROM oauth_accounts WHERE provider = $1 AND provider_user_id = $2`

	oauthGetByUserIDQuery = `
		SELECT id, user_id, provider, provider_user_id, access_token, refresh_token, created_at
		FROM oauth_accounts WHERE user_id = $1`

	oauthUpdateQuery = `
		UPDATE oauth_accounts 
		SET access_token = $2, refresh_token = $3
		WHERE id = $1`

	oauthDeleteQuery = `DELETE FROM oauth_accounts WHERE id = $1`
)
