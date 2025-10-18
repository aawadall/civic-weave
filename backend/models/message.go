package models

import (
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
)

// ProjectMessage represents a message (expanded for universal messaging)
type ProjectMessage struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	ProjectID       *uuid.UUID `json:"project_id,omitempty" db:"project_id"`
	SenderID        uuid.UUID  `json:"sender_id" db:"sender_id"`
	RecipientUserID *uuid.UUID `json:"recipient_user_id,omitempty" db:"recipient_user_id"`
	RecipientTeamID *uuid.UUID `json:"recipient_team_id,omitempty" db:"recipient_team_id"`
	Subject         *string    `json:"subject,omitempty" db:"subject"`
	MessageText     string     `json:"message_text" db:"message_text"`
	TaskID          *uuid.UUID `json:"task_id,omitempty" db:"task_id"`
	MessageType     string     `json:"message_type" db:"message_type"`
	MessageScope    string     `json:"message_scope" db:"message_scope"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	EditedAt        *time.Time `json:"edited_at,omitempty" db:"edited_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
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
	SenderName     string  `json:"sender_name"`
	SenderEmail    string  `json:"sender_email"`
	RecipientName  *string `json:"recipient_name,omitempty"`
	RecipientEmail *string `json:"recipient_email,omitempty"`
	ProjectTitle   *string `json:"project_title,omitempty"`
	IsRead         bool    `json:"is_read"`
}

// Conversation represents a message conversation/thread
type Conversation struct {
	ID               uuid.UUID          `json:"id"`
	Type             string             `json:"type"` // "user_to_user", "user_to_team", "project"
	Title            string             `json:"title"`
	LastMessage      *MessageWithSender `json:"last_message,omitempty"`
	UnreadCount      int                `json:"unread_count"`
	ParticipantCount int                `json:"participant_count"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

// UnreadCount represents unread message count for a project
type UnreadCount struct {
	ProjectID uuid.UUID `json:"project_id"`
	Count     int       `json:"count"`
}

// UniversalUnreadCount represents unread message count across all contexts
type UniversalUnreadCount struct {
	DirectMessages  int `json:"direct_messages"`
	TeamMessages    int `json:"team_messages"`
	ProjectMessages int `json:"project_messages"`
	Total           int `json:"total"`
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

	// Start transaction to create message and auto-record sender read receipt
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert the message
	err = tx.QueryRow(messageCreateQuery, message.ID, message.ProjectID, message.SenderID, message.MessageText, message.TaskID, message.MessageType).
		Scan(&message.CreatedAt)
	if err != nil {
		return err
	}

	// Auto-record read receipt for sender
	_, err = tx.Exec(messageMarkAsReadQuery, message.SenderID, message.ID)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
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
		ProjectID:    &projectID,
		SenderID:     senderID,
		MessageText:  messageText,
		TaskID:       &taskID,
		MessageType:  messageType,
		MessageScope: "project",
	}
	return s.Create(message)
}

// CreateUniversalMessage creates a message for universal messaging
func (s *MessageService) CreateUniversalMessage(message *ProjectMessage) error {
	message.ID = uuid.New()
	// Set default message scope if not specified
	if message.MessageScope == "" {
		if message.RecipientUserID != nil {
			message.MessageScope = "user_to_user"
		} else if message.RecipientTeamID != nil {
			message.MessageScope = "user_to_team"
		} else if message.ProjectID != nil {
			message.MessageScope = "project"
		}
	}
	// Set default message type if not specified
	if message.MessageType == "" {
		message.MessageType = "general"
	}

	subjectStr := "nil"
	if message.Subject != nil {
		subjectStr = *message.Subject
	}
	log.Printf("DEBUG: CreateUniversalMessage - ID: %s, ProjectID: %v, SenderID: %s, RecipientUserID: %v, RecipientTeamID: %v, Subject: %s, MessageText: %s, TaskID: %v, MessageType: %s, MessageScope: %s",
		message.ID, message.ProjectID, message.SenderID, message.RecipientUserID, message.RecipientTeamID, subjectStr, message.MessageText, message.TaskID, message.MessageType, message.MessageScope)

	// Start transaction to create message and auto-record sender read receipt
	tx, err := s.db.Begin()
	if err != nil {
		log.Printf("DEBUG: Failed to begin transaction in CreateUniversalMessage: %v", err)
		return err
	}
	defer tx.Rollback()

	// Insert the message
	err = tx.QueryRow(messageCreateUniversalQuery, message.ID, message.ProjectID, message.SenderID,
		message.RecipientUserID, message.RecipientTeamID, message.Subject, message.MessageText,
		message.TaskID, message.MessageType, message.MessageScope).
		Scan(&message.CreatedAt)

	if err != nil {
		log.Printf("DEBUG: Database error in CreateUniversalMessage: %v", err)
		return err
	}

	// Auto-record read receipt for sender
	_, err = tx.Exec(messageMarkAsReadQuery, message.SenderID, message.ID)
	if err != nil {
		log.Printf("DEBUG: Failed to record sender read receipt in CreateUniversalMessage: %v", err)
		return err
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("DEBUG: Failed to commit transaction in CreateUniversalMessage: %v", err)
		return err
	}

	log.Printf("DEBUG: CreateUniversalMessage successful, CreatedAt: %v", message.CreatedAt)
	return nil
}

// GetInbox retrieves user's inbox (all messages where user is recipient)
func (s *MessageService) GetInbox(userID uuid.UUID, limit, offset int) ([]MessageWithSender, error) {
	rows, err := s.db.Query(messageGetInboxQuery, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []MessageWithSender
	for rows.Next() {
		var msg MessageWithSender
		err := rows.Scan(
			&msg.ID, &msg.ProjectID, &msg.SenderID, &msg.RecipientUserID, &msg.RecipientTeamID,
			&msg.Subject, &msg.MessageText, &msg.TaskID, &msg.MessageType, &msg.MessageScope,
			&msg.CreatedAt, &msg.EditedAt, &msg.DeletedAt,
			&msg.SenderName, &msg.SenderEmail, &msg.RecipientName, &msg.RecipientEmail,
			&msg.ProjectTitle, &msg.IsRead,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// GetSentMessages retrieves messages sent by user
func (s *MessageService) GetSentMessages(userID uuid.UUID, limit, offset int) ([]MessageWithSender, error) {
	rows, err := s.db.Query(messageGetSentQuery, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []MessageWithSender
	for rows.Next() {
		var msg MessageWithSender
		err := rows.Scan(
			&msg.ID, &msg.ProjectID, &msg.SenderID, &msg.RecipientUserID, &msg.RecipientTeamID,
			&msg.Subject, &msg.MessageText, &msg.TaskID, &msg.MessageType, &msg.MessageScope,
			&msg.CreatedAt, &msg.EditedAt, &msg.DeletedAt,
			&msg.SenderName, &msg.SenderEmail, &msg.RecipientName, &msg.RecipientEmail,
			&msg.ProjectTitle, &msg.IsRead,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// GetConversations retrieves user's conversations grouped by thread
func (s *MessageService) GetConversations(userID uuid.UUID, limit, offset int) ([]Conversation, error) {
	rows, err := s.db.Query(messageGetConversationsQuery, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var conv Conversation
		var lastMessageID *uuid.UUID
		var lastMessageText *string
		var lastMessageCreatedAt *time.Time
		var senderName *string
		var senderEmail *string

		err := rows.Scan(
			&conv.ID, &conv.Type, &conv.Title, &conv.UnreadCount, &conv.ParticipantCount,
			&conv.UpdatedAt, &lastMessageID, &lastMessageText, &lastMessageCreatedAt,
			&senderName, &senderEmail,
		)
		if err != nil {
			return nil, err
		}

		// Build last message if exists
		if lastMessageID != nil {
			conv.LastMessage = &MessageWithSender{
				ProjectMessage: ProjectMessage{
					ID:          *lastMessageID,
					MessageText: *lastMessageText,
					CreatedAt:   *lastMessageCreatedAt,
				},
				SenderName:  *senderName,
				SenderEmail: *senderEmail,
			}
		}

		conversations = append(conversations, conv)
	}

	return conversations, rows.Err()
}

// GetConversation retrieves messages in a specific conversation
func (s *MessageService) GetConversation(conversationID uuid.UUID, userID uuid.UUID, limit, offset int) ([]MessageWithSender, error) {
	rows, err := s.db.Query(messageGetConversationQuery, conversationID, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []MessageWithSender
	for rows.Next() {
		var msg MessageWithSender
		err := rows.Scan(
			&msg.ID, &msg.ProjectID, &msg.SenderID, &msg.RecipientUserID, &msg.RecipientTeamID,
			&msg.Subject, &msg.MessageText, &msg.TaskID, &msg.MessageType, &msg.MessageScope,
			&msg.CreatedAt, &msg.EditedAt, &msg.DeletedAt,
			&msg.SenderName, &msg.SenderEmail, &msg.RecipientName, &msg.RecipientEmail,
			&msg.ProjectTitle, &msg.IsRead,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// GetUniversalUnreadCount returns unread message counts across all contexts
func (s *MessageService) GetUniversalUnreadCount(userID uuid.UUID) (*UniversalUnreadCount, error) {
	count := &UniversalUnreadCount{}
	err := s.db.QueryRow(messageGetUniversalUnreadCountQuery, userID).Scan(
		&count.DirectMessages, &count.TeamMessages, &count.ProjectMessages, &count.Total,
	)
	if err != nil {
		return nil, err
	}
	return count, nil
}

// SearchUser represents a user search result
type SearchUser struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Type  string    `json:"type"`
}

// SearchProject represents a project search result
type SearchProject struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title"`
	Type  string    `json:"type"`
}

// SearchUsers searches for users by name or email
func (s *MessageService) SearchUsers(query string, limit int) ([]SearchUser, error) {
	rows, err := s.db.Query(messageSearchUsersQuery, query, limit)
	if err != nil {
		return []SearchUser{}, err
	}
	defer rows.Close()

	var users []SearchUser
	for rows.Next() {
		var user SearchUser
		err := rows.Scan(&user.ID, &user.Name, &user.Email)
		if err != nil {
			return []SearchUser{}, err
		}
		user.Type = "user"
		users = append(users, user)
	}

	return users, rows.Err()
}

// SearchUserProjects searches for projects where user is enrolled
func (s *MessageService) SearchUserProjects(userID uuid.UUID, query string, limit int) ([]SearchProject, error) {
	rows, err := s.db.Query(messageSearchUserProjectsQuery, userID, query, limit)
	if err != nil {
		return []SearchProject{}, err
	}
	defer rows.Close()

	var projects []SearchProject
	for rows.Next() {
		var project SearchProject
		err := rows.Scan(&project.ID, &project.Title)
		if err != nil {
			return []SearchProject{}, err
		}
		project.Type = "project"
		projects = append(projects, project)
	}

	return projects, rows.Err()
}
