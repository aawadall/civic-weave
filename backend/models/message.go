package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// ProjectMessage represents a message in a project
type ProjectMessage struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ProjectID   uuid.UUID  `json:"project_id" db:"project_id"`
	SenderID    uuid.UUID  `json:"sender_id" db:"sender_id"`
	MessageText string     `json:"message_text" db:"message_text"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	EditedAt    *time.Time `json:"edited_at,omitempty" db:"edited_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// MessageRead represents a read receipt for a message
type MessageRead struct {
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	MessageID uuid.UUID `json:"message_id" db:"message_id"`
	ReadAt    time.Time `json:"read_at" db:"read_at"`
}

// MessageWithSender includes message and sender info
type MessageWithSender struct {
	ProjectMessage
	SenderName  string `json:"sender_name"`
	SenderEmail string `json:"sender_email"`
	IsRead      bool   `json:"is_read"`
}

// UnreadCount represents unread message count for a project
type UnreadCount struct {
	ProjectID uuid.UUID `json:"project_id"`
	Count     int       `json:"count"`
}

// MessageService handles message operations
type MessageService struct {
	db *sql.DB
}

// NewMessageService creates a new message service
func NewMessageService(db *sql.DB) *MessageService {
	return &MessageService{db: db}
}

// Create creates a new message
func (s *MessageService) Create(message *ProjectMessage) error {
	query := `
		INSERT INTO project_messages (id, project_id, sender_id, message_text)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`

	message.ID = uuid.New()
	return s.db.QueryRow(query, message.ID, message.ProjectID, message.SenderID, message.MessageText).
		Scan(&message.CreatedAt)
}

// GetByID retrieves a message by ID
func (s *MessageService) GetByID(id uuid.UUID) (*ProjectMessage, error) {
	message := &ProjectMessage{}
	query := `
		SELECT id, project_id, sender_id, message_text, created_at, edited_at, deleted_at
		FROM project_messages WHERE id = $1`

	err := s.db.QueryRow(query, id).Scan(
		&message.ID, &message.ProjectID, &message.SenderID, &message.MessageText,
		&message.CreatedAt, &message.EditedAt, &message.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return message, nil
}

// ListByProject retrieves messages for a project (paginated)
// Excludes soft-deleted messages
func (s *MessageService) ListByProject(projectID uuid.UUID, limit, offset int, userID *uuid.UUID) ([]MessageWithSender, error) {
	query := `
		SELECT 
			pm.id, pm.project_id, pm.sender_id, pm.message_text, 
			pm.created_at, pm.edited_at, pm.deleted_at,
			u.email as sender_email,
			COALESCE(v.name, a.name, u.email) as sender_name,
			CASE WHEN mr.user_id IS NOT NULL THEN true ELSE false END as is_read
		FROM project_messages pm
		JOIN users u ON pm.sender_id = u.id
		LEFT JOIN volunteers v ON u.id = v.user_id
		LEFT JOIN admins a ON u.id = a.user_id
		LEFT JOIN message_reads mr ON pm.id = mr.message_id AND mr.user_id = $4
		WHERE pm.project_id = $1 AND pm.deleted_at IS NULL
		ORDER BY pm.created_at ASC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(query, projectID, limit, offset, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []MessageWithSender
	for rows.Next() {
		var msg MessageWithSender
		err := rows.Scan(
			&msg.ID, &msg.ProjectID, &msg.SenderID, &msg.MessageText,
			&msg.CreatedAt, &msg.EditedAt, &msg.DeletedAt,
			&msg.SenderEmail, &msg.SenderName, &msg.IsRead,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// ListRecentByProject retrieves recent messages for a project (last N messages)
func (s *MessageService) ListRecentByProject(projectID uuid.UUID, count int, userID *uuid.UUID) ([]MessageWithSender, error) {
	query := `
		SELECT 
			pm.id, pm.project_id, pm.sender_id, pm.message_text, 
			pm.created_at, pm.edited_at, pm.deleted_at,
			u.email as sender_email,
			COALESCE(v.name, a.name, u.email) as sender_name,
			CASE WHEN mr.user_id IS NOT NULL THEN true ELSE false END as is_read
		FROM project_messages pm
		JOIN users u ON pm.sender_id = u.id
		LEFT JOIN volunteers v ON u.id = v.user_id
		LEFT JOIN admins a ON u.id = a.user_id
		LEFT JOIN message_reads mr ON pm.id = mr.message_id AND mr.user_id = $3
		WHERE pm.project_id = $1 AND pm.deleted_at IS NULL
		ORDER BY pm.created_at DESC
		LIMIT $2`

	rows, err := s.db.Query(query, projectID, count, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []MessageWithSender
	for rows.Next() {
		var msg MessageWithSender
		err := rows.Scan(
			&msg.ID, &msg.ProjectID, &msg.SenderID, &msg.MessageText,
			&msg.CreatedAt, &msg.EditedAt, &msg.DeletedAt,
			&msg.SenderEmail, &msg.SenderName, &msg.IsRead,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, rows.Err()
}

// Update updates a message (for editing)
func (s *MessageService) Update(message *ProjectMessage) error {
	query := `
		UPDATE project_messages 
		SET message_text = $2, edited_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING edited_at`

	return s.db.QueryRow(query, message.ID, message.MessageText).Scan(&message.EditedAt)
}

// SoftDelete soft-deletes a message
func (s *MessageService) SoftDelete(id uuid.UUID) error {
	query := `
		UPDATE project_messages 
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// MarkAsRead marks a message as read for a user
func (s *MessageService) MarkAsRead(messageID, userID uuid.UUID) error {
	query := `
		INSERT INTO message_reads (user_id, message_id, read_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, message_id) DO NOTHING`

	_, err := s.db.Exec(query, userID, messageID)
	return err
}

// MarkAllAsRead marks all messages in a project as read for a user
func (s *MessageService) MarkAllAsRead(projectID, userID uuid.UUID) error {
	query := `
		INSERT INTO message_reads (user_id, message_id, read_at)
		SELECT $1, pm.id, CURRENT_TIMESTAMP
		FROM project_messages pm
		WHERE pm.project_id = $2 AND pm.deleted_at IS NULL
		ON CONFLICT (user_id, message_id) DO NOTHING`

	_, err := s.db.Exec(query, userID, projectID)
	return err
}

// GetUnreadCount returns the count of unread messages for a user in a project
func (s *MessageService) GetUnreadCount(projectID, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM project_messages pm
		LEFT JOIN message_reads mr ON pm.id = mr.message_id AND mr.user_id = $2
		WHERE pm.project_id = $1 
		  AND pm.deleted_at IS NULL
		  AND pm.sender_id != $2
		  AND mr.user_id IS NULL`

	var count int
	err := s.db.QueryRow(query, projectID, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetUnreadCountsByUser returns unread message counts for all projects a user is enrolled in
func (s *MessageService) GetUnreadCountsByUser(userID uuid.UUID) ([]UnreadCount, error) {
	query := `
		SELECT 
			pm.project_id,
			COUNT(*) as unread_count
		FROM project_messages pm
		JOIN project_team_members ptm ON pm.project_id = ptm.project_id
		LEFT JOIN message_reads mr ON pm.id = mr.message_id AND mr.user_id = $1
		WHERE ptm.volunteer_id = (
			SELECT id FROM volunteers WHERE user_id = $1
		)
		  AND ptm.status = 'active'
		  AND pm.deleted_at IS NULL
		  AND pm.sender_id != $1
		  AND mr.user_id IS NULL
		GROUP BY pm.project_id`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []UnreadCount
	for rows.Next() {
		var count UnreadCount
		err := rows.Scan(&count.ProjectID, &count.Count)
		if err != nil {
			return nil, err
		}
		counts = append(counts, count)
	}

	return counts, rows.Err()
}

// GetMessagesAfter retrieves messages created after a specific timestamp (for polling)
func (s *MessageService) GetMessagesAfter(projectID uuid.UUID, after time.Time, userID *uuid.UUID) ([]MessageWithSender, error) {
	query := `
		SELECT 
			pm.id, pm.project_id, pm.sender_id, pm.message_text, 
			pm.created_at, pm.edited_at, pm.deleted_at,
			u.email as sender_email,
			COALESCE(v.name, a.name, u.email) as sender_name,
			CASE WHEN mr.user_id IS NOT NULL THEN true ELSE false END as is_read
		FROM project_messages pm
		JOIN users u ON pm.sender_id = u.id
		LEFT JOIN volunteers v ON u.id = v.user_id
		LEFT JOIN admins a ON u.id = a.user_id
		LEFT JOIN message_reads mr ON pm.id = mr.message_id AND mr.user_id = $3
		WHERE pm.project_id = $1 
		  AND pm.created_at > $2 
		  AND pm.deleted_at IS NULL
		ORDER BY pm.created_at ASC`

	rows, err := s.db.Query(query, projectID, after, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []MessageWithSender
	for rows.Next() {
		var msg MessageWithSender
		err := rows.Scan(
			&msg.ID, &msg.ProjectID, &msg.SenderID, &msg.MessageText,
			&msg.CreatedAt, &msg.EditedAt, &msg.DeletedAt,
			&msg.SenderEmail, &msg.SenderName, &msg.IsRead,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// CanUserEdit checks if a user can edit a message (sender and within 15 minutes)
func (s *MessageService) CanUserEdit(messageID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT 
			CASE 
				WHEN sender_id = $2 AND created_at > (CURRENT_TIMESTAMP - INTERVAL '15 minutes') 
				THEN true 
				ELSE false 
			END
		FROM project_messages
		WHERE id = $1`

	var canEdit bool
	err := s.db.QueryRow(query, messageID, userID).Scan(&canEdit)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return canEdit, nil
}

