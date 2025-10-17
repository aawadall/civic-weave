package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// BroadcastMessage represents a system-wide announcement
type BroadcastMessage struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Title          string     `json:"title" db:"title"`
	Content        string     `json:"content" db:"content"`
	AuthorID       uuid.UUID  `json:"author_id" db:"author_id"`
	TargetAudience string     `json:"target_audience" db:"target_audience"`
	Priority       string     `json:"priority" db:"priority"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// BroadcastRead represents a read receipt for a broadcast
type BroadcastRead struct {
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	BroadcastID uuid.UUID `json:"broadcast_id" db:"broadcast_id"`
	ReadAt      time.Time `json:"read_at" db:"read_at"`
}

// BroadcastWithAuthor includes broadcast and author info
type BroadcastWithAuthor struct {
	BroadcastMessage
	AuthorName  string `json:"author_name"`
	AuthorEmail string `json:"author_email"`
	IsRead      bool   `json:"is_read"`
}

// BroadcastStats represents statistics for broadcasts
type BroadcastStats struct {
	TotalBroadcasts   int `json:"total_broadcasts"`
	UnreadBroadcasts  int `json:"unread_broadcasts"`
	HighPriorityCount int `json:"high_priority_count"`
	UrgentCount       int `json:"urgent_count"`
}

// BroadcastService handles broadcast operations
type BroadcastService struct {
	db *sql.DB
}

// NewBroadcastService creates a new broadcast service
func NewBroadcastService(db *sql.DB) *BroadcastService {
	return &BroadcastService{db: db}
}

// Create creates a new broadcast
func (s *BroadcastService) Create(broadcast *BroadcastMessage) error {
	broadcast.ID = uuid.New()
	return s.db.QueryRow(broadcastCreateQuery, broadcast.ID, broadcast.Title, broadcast.Content,
		broadcast.AuthorID, broadcast.TargetAudience, broadcast.Priority, broadcast.ExpiresAt).
		Scan(&broadcast.CreatedAt, &broadcast.UpdatedAt)
}

// GetByID retrieves a broadcast by ID
func (s *BroadcastService) GetByID(id uuid.UUID) (*BroadcastMessage, error) {
	broadcast := &BroadcastMessage{}
	err := s.db.QueryRow(broadcastGetByIDQuery, id).Scan(
		&broadcast.ID, &broadcast.Title, &broadcast.Content, &broadcast.AuthorID,
		&broadcast.TargetAudience, &broadcast.Priority, &broadcast.ExpiresAt,
		&broadcast.CreatedAt, &broadcast.UpdatedAt, &broadcast.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return broadcast, nil
}

// List retrieves broadcasts for a user based on their role and target audience
func (s *BroadcastService) List(userID uuid.UUID, userRole string, limit, offset int) ([]BroadcastWithAuthor, error) {
	rows, err := s.db.Query(broadcastListQuery, userID, userRole, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var broadcasts []BroadcastWithAuthor
	for rows.Next() {
		var broadcast BroadcastWithAuthor
		err := rows.Scan(
			&broadcast.ID, &broadcast.Title, &broadcast.Content, &broadcast.AuthorID,
			&broadcast.TargetAudience, &broadcast.Priority, &broadcast.ExpiresAt,
			&broadcast.CreatedAt, &broadcast.UpdatedAt, &broadcast.DeletedAt,
			&broadcast.AuthorName, &broadcast.AuthorEmail, &broadcast.IsRead,
		)
		if err != nil {
			return nil, err
		}
		broadcasts = append(broadcasts, broadcast)
	}

	return broadcasts, rows.Err()
}

// ListAll retrieves all broadcasts (admin only)
func (s *BroadcastService) ListAll(limit, offset int) ([]BroadcastWithAuthor, error) {
	rows, err := s.db.Query(broadcastListAllQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var broadcasts []BroadcastWithAuthor
	for rows.Next() {
		var broadcast BroadcastWithAuthor
		err := rows.Scan(
			&broadcast.ID, &broadcast.Title, &broadcast.Content, &broadcast.AuthorID,
			&broadcast.TargetAudience, &broadcast.Priority, &broadcast.ExpiresAt,
			&broadcast.CreatedAt, &broadcast.UpdatedAt, &broadcast.DeletedAt,
			&broadcast.AuthorName, &broadcast.AuthorEmail, &broadcast.IsRead,
		)
		if err != nil {
			return nil, err
		}
		broadcasts = append(broadcasts, broadcast)
	}

	return broadcasts, rows.Err()
}

// Update updates a broadcast
func (s *BroadcastService) Update(broadcast *BroadcastMessage) error {
	return s.db.QueryRow(broadcastUpdateQuery, broadcast.ID, broadcast.Title, broadcast.Content,
		broadcast.TargetAudience, broadcast.Priority, broadcast.ExpiresAt).
		Scan(&broadcast.UpdatedAt)
}

// SoftDelete soft-deletes a broadcast
func (s *BroadcastService) SoftDelete(id uuid.UUID) error {
	result, err := s.db.Exec(broadcastSoftDeleteQuery, id)
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

// MarkAsRead marks a broadcast as read for a user
func (s *BroadcastService) MarkAsRead(broadcastID, userID uuid.UUID) error {
	_, err := s.db.Exec(broadcastMarkAsReadQuery, userID, broadcastID)
	return err
}

// GetUnreadCount returns the count of unread broadcasts for a user
func (s *BroadcastService) GetUnreadCount(userID uuid.UUID, userRoles []string) (int, error) {
	var count int
	err := s.db.QueryRow(broadcastGetUnreadCountQuery, userID, userRoles).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetStats returns broadcast statistics for a user
func (s *BroadcastService) GetStats(userID uuid.UUID, userRoles []string) (*BroadcastStats, error) {
	stats := &BroadcastStats{}
	err := s.db.QueryRow(broadcastGetStatsQuery, userID, userRoles).Scan(
		&stats.TotalBroadcasts, &stats.UnreadBroadcasts, &stats.HighPriorityCount, &stats.UrgentCount,
	)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// IsAuthor checks if a user is the author of a broadcast
func (s *BroadcastService) IsAuthor(broadcastID, userID uuid.UUID) (bool, error) {
	var count int
	err := s.db.QueryRow(broadcastIsAuthorQuery, broadcastID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetDB returns the database connection (needed for creating other services)
func (s *BroadcastService) GetDB() *sql.DB {
	return s.db
}
