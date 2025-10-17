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
	TaskID      *uuid.UUID `json:"task_id,omitempty" db:"task_id"`
	MessageType string     `json:"message_type" db:"message_type"`
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
	message.ID = uuid.New()
	// Set default message type if not specified
	if message.MessageType == "" {
		message.MessageType = "general"
	}
	return s.db.QueryRow(messageCreateQuery, message.ID, message.ProjectID, message.SenderID, message.MessageText, message.TaskID, message.MessageType).
		Scan(&message.CreatedAt)
}

// GetByID retrieves a message by ID
func (s *MessageService) GetByID(id uuid.UUID) (*ProjectMessage, error) {
	message := &ProjectMessage{}

	err := s.db.QueryRow(messageGetByIDQuery, id).Scan(
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
	rows, err := s.db.Query(messageListByProjectQuery, projectID, limit, offset, userID)
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
	rows, err := s.db.Query(messageListRecentByProjectQuery, projectID, count, userID)
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
	return s.db.QueryRow(messageUpdateQuery, message.ID, message.MessageText).Scan(&message.EditedAt)
}

// SoftDelete soft-deletes a message
func (s *MessageService) SoftDelete(id uuid.UUID) error {
	result, err := s.db.Exec(messageSoftDeleteQuery, id)
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
	_, err := s.db.Exec(messageMarkAsReadQuery, userID, messageID)
	return err
}

// MarkAllAsRead marks all messages in a project as read for a user
func (s *MessageService) MarkAllAsRead(projectID, userID uuid.UUID) error {
	_, err := s.db.Exec(messageMarkAllAsReadQuery, userID, projectID)
	return err
}

// GetUnreadCount returns the count of unread messages for a user in a project
func (s *MessageService) GetUnreadCount(projectID, userID uuid.UUID) (int, error) {
	var count int
	err := s.db.QueryRow(messageGetUnreadCountQuery, projectID, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetUnreadCountsByUser returns unread message counts for all projects a user is enrolled in
func (s *MessageService) GetUnreadCountsByUser(userID uuid.UUID) ([]UnreadCount, error) {
	rows, err := s.db.Query(messageGetUnreadCountsByUserQuery, userID)
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
	rows, err := s.db.Query(messageGetMessagesAfterQuery, projectID, after, userID)
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
	var canEdit bool
	err := s.db.QueryRow(messageCanUserEditQuery, messageID, userID).Scan(&canEdit)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return canEdit, nil
}

// CreateTaskNotification creates a task-related notification message
func (s *MessageService) CreateTaskNotification(projectID, senderID, taskID uuid.UUID, messageType, messageText string) error {
	message := &ProjectMessage{
		ProjectID:   projectID,
		SenderID:    senderID,
		MessageText: messageText,
		TaskID:      &taskID,
		MessageType: messageType,
	}
	return s.Create(message)
}
