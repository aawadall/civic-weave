package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// CampaignStatus represents the status of a campaign
type CampaignStatus string

const (
	CampaignStatusDraft     CampaignStatus = "draft"
	CampaignStatusScheduled CampaignStatus = "scheduled"
	CampaignStatusSending   CampaignStatus = "sending"
	CampaignStatusSent      CampaignStatus = "sent"
	CampaignStatusFailed    CampaignStatus = "failed"
)

// RecipientStatus represents the status of a campaign recipient
type RecipientStatus string

const (
	RecipientStatusPending   RecipientStatus = "pending"
	RecipientStatusSent      RecipientStatus = "sent"
	RecipientStatusDelivered RecipientStatus = "delivered"
	RecipientStatusOpened    RecipientStatus = "opened"
	RecipientStatusClicked   RecipientStatus = "clicked"
	RecipientStatusFailed    RecipientStatus = "failed"
)

// Campaign represents an email campaign
type Campaign struct {
	ID              uuid.UUID      `json:"id" db:"id"`
	Title           string         `json:"title" db:"title"`
	Description     string         `json:"description" db:"description"`
	TargetRoles     []string       `json:"target_roles" db:"target_roles"`
	Status          CampaignStatus `json:"status" db:"status"`
	EmailSubject    string         `json:"email_subject" db:"email_subject"`
	EmailBody       string         `json:"email_body" db:"email_body"`
	CreatedByUserID uuid.UUID      `json:"created_by_user_id" db:"created_by_user_id"`
	ScheduledAt     *time.Time     `json:"scheduled_at" db:"scheduled_at"`
	SentAt          *time.Time     `json:"sent_at" db:"sent_at"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
}

// CampaignRecipient represents a campaign recipient
type CampaignRecipient struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	CampaignID uuid.UUID       `json:"campaign_id" db:"campaign_id"`
	UserID     uuid.UUID       `json:"user_id" db:"user_id"`
	SentAt     *time.Time      `json:"sent_at" db:"sent_at"`
	OpenedAt   *time.Time      `json:"opened_at" db:"opened_at"`
	ClickedAt  *time.Time      `json:"clicked_at" db:"clicked_at"`
	Status     RecipientStatus `json:"status" db:"status"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

// CampaignWithStats represents a campaign with delivery statistics
type CampaignWithStats struct {
	Campaign
	TotalRecipients int     `json:"total_recipients"`
	SentCount       int     `json:"sent_count"`
	DeliveredCount  int     `json:"delivered_count"`
	OpenedCount     int     `json:"opened_count"`
	ClickedCount    int     `json:"clicked_count"`
	FailedCount     int     `json:"failed_count"`
	DeliveryRate    float64 `json:"delivery_rate"`
	OpenRate        float64 `json:"open_rate"`
	ClickRate       float64 `json:"click_rate"`
}

// CampaignService handles campaign operations
type CampaignService struct {
	db *sql.DB
}

// NewCampaignService creates a new campaign service
func NewCampaignService(db *sql.DB) *CampaignService {
	return &CampaignService{db: db}
}

// CreateCampaign creates a new campaign
func (s *CampaignService) CreateCampaign(campaign *Campaign) error {
	query := `
		INSERT INTO campaigns (id, title, description, target_roles, status, email_subject, email_body, created_by_user_id, scheduled_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at`

	campaign.ID = uuid.New()
	targetRolesJSON, err := ToJSONArray(campaign.TargetRoles)
	if err != nil {
		return err
	}

	return s.db.QueryRow(query, campaign.ID, campaign.Title, campaign.Description,
		targetRolesJSON, campaign.Status, campaign.EmailSubject, campaign.EmailBody,
		campaign.CreatedByUserID, campaign.ScheduledAt).
		Scan(&campaign.CreatedAt, &campaign.UpdatedAt)
}

// GetCampaignByID retrieves a campaign by ID
func (s *CampaignService) GetCampaignByID(id uuid.UUID) (*Campaign, error) {
	campaign := &Campaign{}
	var targetRolesJSON string
	query := `
		SELECT id, title, description, target_roles, status, email_subject, email_body, 
		       created_by_user_id, scheduled_at, sent_at, created_at, updated_at
		FROM campaigns WHERE id = $1`

	err := s.db.QueryRow(query, id).Scan(&campaign.ID, &campaign.Title, &campaign.Description,
		&targetRolesJSON, &campaign.Status, &campaign.EmailSubject, &campaign.EmailBody,
		&campaign.CreatedByUserID, &campaign.ScheduledAt, &campaign.SentAt,
		&campaign.CreatedAt, &campaign.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse target roles JSON
	if err := ParseJSONArray(targetRolesJSON, &campaign.TargetRoles); err != nil {
		return nil, err
	}

	return campaign, nil
}

// ListCampaigns retrieves all campaigns with optional filtering
func (s *CampaignService) ListCampaigns(limit, offset int, status *CampaignStatus, createdByUserID *uuid.UUID) ([]Campaign, error) {
	query := `
		SELECT id, title, description, target_roles, status, email_subject, email_body, 
		       created_by_user_id, scheduled_at, sent_at, created_at, updated_at
		FROM campaigns
		WHERE ($1 IS NULL OR status = $1)
		AND ($2 IS NULL OR created_by_user_id = $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := s.db.Query(query, status, createdByUserID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []Campaign
	for rows.Next() {
		var campaign Campaign
		var targetRolesJSON string
		err := rows.Scan(&campaign.ID, &campaign.Title, &campaign.Description,
			&targetRolesJSON, &campaign.Status, &campaign.EmailSubject, &campaign.EmailBody,
			&campaign.CreatedByUserID, &campaign.ScheduledAt, &campaign.SentAt,
			&campaign.CreatedAt, &campaign.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// Parse target roles JSON
		if err := ParseJSONArray(targetRolesJSON, &campaign.TargetRoles); err != nil {
			return nil, err
		}

		campaigns = append(campaigns, campaign)
	}

	return campaigns, nil
}

// UpdateCampaign updates a campaign
func (s *CampaignService) UpdateCampaign(campaign *Campaign) error {
	query := `
		UPDATE campaigns 
		SET title = $2, description = $3, target_roles = $4, status = $5, 
		    email_subject = $6, email_body = $7, scheduled_at = $8, sent_at = $9,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	targetRolesJSON, err := ToJSONArray(campaign.TargetRoles)
	if err != nil {
		return err
	}

	return s.db.QueryRow(query, campaign.ID, campaign.Title, campaign.Description,
		targetRolesJSON, campaign.Status, campaign.EmailSubject, campaign.EmailBody,
		campaign.ScheduledAt, campaign.SentAt).Scan(&campaign.UpdatedAt)
}

// DeleteCampaign deletes a campaign
func (s *CampaignService) DeleteCampaign(id uuid.UUID) error {
	query := `DELETE FROM campaigns WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}

// GetCampaignRecipients retrieves all recipients for a campaign
func (s *CampaignService) GetCampaignRecipients(campaignID uuid.UUID) ([]CampaignRecipient, error) {
	query := `
		SELECT id, campaign_id, user_id, sent_at, opened_at, clicked_at, status, created_at
		FROM campaign_recipients 
		WHERE campaign_id = $1
		ORDER BY created_at`

	rows, err := s.db.Query(query, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipients []CampaignRecipient
	for rows.Next() {
		var recipient CampaignRecipient
		err := rows.Scan(&recipient.ID, &recipient.CampaignID, &recipient.UserID,
			&recipient.SentAt, &recipient.OpenedAt, &recipient.ClickedAt,
			&recipient.Status, &recipient.CreatedAt)
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, recipient)
	}

	return recipients, nil
}

// GetCampaignStats retrieves delivery statistics for a campaign
func (s *CampaignService) GetCampaignStats(campaignID uuid.UUID) (*CampaignWithStats, error) {
	campaign, err := s.GetCampaignByID(campaignID)
	if err != nil {
		return nil, err
	}
	if campaign == nil {
		return nil, nil
	}

	stats := &CampaignWithStats{
		Campaign: *campaign,
	}

	// Get recipient statistics
	query := `
		SELECT 
			COUNT(*) as total_recipients,
			COUNT(CASE WHEN status IN ('sent', 'delivered', 'opened', 'clicked') THEN 1 END) as sent_count,
			COUNT(CASE WHEN status IN ('delivered', 'opened', 'clicked') THEN 1 END) as delivered_count,
			COUNT(CASE WHEN status IN ('opened', 'clicked') THEN 1 END) as opened_count,
			COUNT(CASE WHEN status = 'clicked' THEN 1 END) as clicked_count,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_count
		FROM campaign_recipients 
		WHERE campaign_id = $1`

	err = s.db.QueryRow(query, campaignID).Scan(
		&stats.TotalRecipients, &stats.SentCount, &stats.DeliveredCount,
		&stats.OpenedCount, &stats.ClickedCount, &stats.FailedCount)
	if err != nil {
		return nil, err
	}

	// Calculate rates
	if stats.TotalRecipients > 0 {
		stats.DeliveryRate = float64(stats.DeliveredCount) / float64(stats.TotalRecipients)
		if stats.DeliveredCount > 0 {
			stats.OpenRate = float64(stats.OpenedCount) / float64(stats.DeliveredCount)
			if stats.OpenedCount > 0 {
				stats.ClickRate = float64(stats.ClickedCount) / float64(stats.OpenedCount)
			}
		}
	}

	return stats, nil
}

// AddCampaignRecipient adds a recipient to a campaign
func (s *CampaignService) AddCampaignRecipient(campaignID, userID uuid.UUID) error {
	query := `
		INSERT INTO campaign_recipients (id, campaign_id, user_id, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (campaign_id, user_id) DO NOTHING`

	_, err := s.db.Exec(query, uuid.New(), campaignID, userID, RecipientStatusPending)
	return err
}

// UpdateRecipientStatus updates the status of a campaign recipient
func (s *CampaignService) UpdateRecipientStatus(recipientID uuid.UUID, status RecipientStatus) error {
	query := `
		UPDATE campaign_recipients 
		SET status = $2
		WHERE id = $1`

	_, err := s.db.Exec(query, recipientID, status)
	return err
}

// MarkRecipientAsSent marks a recipient as sent
func (s *CampaignService) MarkRecipientAsSent(recipientID uuid.UUID) error {
	query := `
		UPDATE campaign_recipients 
		SET status = 'sent', sent_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	_, err := s.db.Exec(query, recipientID)
	return err
}

// MarkRecipientAsDelivered marks a recipient as delivered
func (s *CampaignService) MarkRecipientAsDelivered(recipientID uuid.UUID) error {
	query := `
		UPDATE campaign_recipients 
		SET status = 'delivered'
		WHERE id = $1 AND status = 'sent'`

	_, err := s.db.Exec(query, recipientID)
	return err
}

// MarkRecipientAsOpened marks a recipient as opened
func (s *CampaignService) MarkRecipientAsOpened(recipientID uuid.UUID) error {
	query := `
		UPDATE campaign_recipients 
		SET status = 'opened', opened_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status IN ('sent', 'delivered')`

	_, err := s.db.Exec(query, recipientID)
	return err
}

// MarkRecipientAsClicked marks a recipient as clicked
func (s *CampaignService) MarkRecipientAsClicked(recipientID uuid.UUID) error {
	query := `
		UPDATE campaign_recipients 
		SET status = 'clicked', clicked_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status IN ('sent', 'delivered', 'opened')`

	_, err := s.db.Exec(query, recipientID)
	return err
}

// GetTargetUsersForCampaign retrieves users who match the campaign's target roles
func (s *CampaignService) GetTargetUsersForCampaign(targetRoles []string) ([]User, error) {
	if len(targetRoles) == 0 {
		return []User{}, nil
	}

	query := `
		SELECT DISTINCT u.id, u.email, u.password_hash, u.email_verified, u.created_at, u.updated_at
		FROM users u
		INNER JOIN user_roles ur ON u.id = ur.user_id
		INNER JOIN roles r ON ur.role_id = r.id
		WHERE r.name = ANY($1)
		AND u.email_verified = true
		ORDER BY u.email`

	rows, err := s.db.Query(query, targetRoles)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.EmailVerified,
			&user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// ScheduleCampaign schedules a campaign for future sending
func (s *CampaignService) ScheduleCampaign(campaignID uuid.UUID, scheduledAt time.Time) error {
	query := `
		UPDATE campaigns 
		SET status = 'scheduled', scheduled_at = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status = 'draft'`

	result, err := s.db.Exec(query, campaignID, scheduledAt)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows // Campaign not found or not in draft status
	}

	return nil
}

// MarkCampaignAsSent marks a campaign as sent
func (s *CampaignService) MarkCampaignAsSent(campaignID uuid.UUID) error {
	query := `
		UPDATE campaigns 
		SET status = 'sent', sent_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	_, err := s.db.Exec(query, campaignID)
	return err
}

// GetScheduledCampaigns retrieves campaigns that are scheduled to be sent
func (s *CampaignService) GetScheduledCampaigns() ([]Campaign, error) {
	query := `
		SELECT id, title, description, target_roles, status, email_subject, email_body, 
		       created_by_user_id, scheduled_at, sent_at, created_at, updated_at
		FROM campaigns
		WHERE status = 'scheduled' AND scheduled_at <= CURRENT_TIMESTAMP
		ORDER BY scheduled_at`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []Campaign
	for rows.Next() {
		var campaign Campaign
		var targetRolesJSON string
		err := rows.Scan(&campaign.ID, &campaign.Title, &campaign.Description,
			&targetRolesJSON, &campaign.Status, &campaign.EmailSubject, &campaign.EmailBody,
			&campaign.CreatedByUserID, &campaign.ScheduledAt, &campaign.SentAt,
			&campaign.CreatedAt, &campaign.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// Parse target roles JSON
		if err := ParseJSONArray(targetRolesJSON, &campaign.TargetRoles); err != nil {
			return nil, err
		}

		campaigns = append(campaigns, campaign)
	}

	return campaigns, nil
}
