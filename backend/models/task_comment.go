package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// TaskComment represents a comment on a task
type TaskComment struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	TaskID      uuid.UUID  `json:"task_id" db:"task_id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	CommentText string     `json:"comment_text" db:"comment_text"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	EditedAt    *time.Time `json:"edited_at,omitempty" db:"edited_at"`
}

// TaskCommentWithUser includes user information
type TaskCommentWithUser struct {
	TaskComment
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
}

// TaskCommentService handles task comment operations
type TaskCommentService struct {
	db *sql.DB
}

// NewTaskCommentService creates a new task comment service
func NewTaskCommentService(db *sql.DB) *TaskCommentService {
	return &TaskCommentService{db: db}
}

// Create creates a new task comment
func (s *TaskCommentService) Create(comment *TaskComment) error {
	comment.ID = uuid.New()
	return s.db.QueryRow(commentCreateQuery, comment.ID, comment.TaskID, comment.UserID, comment.CommentText).
		Scan(&comment.CreatedAt)
}

// GetByTask retrieves all comments for a specific task
func (s *TaskCommentService) GetByTask(taskID uuid.UUID) ([]TaskCommentWithUser, error) {
	rows, err := s.db.Query(commentGetByTaskQuery, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []TaskCommentWithUser
	for rows.Next() {
		var comment TaskCommentWithUser
		err := rows.Scan(
			&comment.ID, &comment.TaskID, &comment.UserID, &comment.CommentText,
			&comment.CreatedAt, &comment.EditedAt, &comment.UserName, &comment.UserEmail,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, rows.Err()
}

// GetByID retrieves a comment by ID
func (s *TaskCommentService) GetByID(id uuid.UUID) (*TaskComment, error) {
	comment := &TaskComment{}
	err := s.db.QueryRow(commentGetByIDQuery, id).Scan(
		&comment.ID, &comment.TaskID, &comment.UserID, &comment.CommentText,
		&comment.CreatedAt, &comment.EditedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return comment, nil
}

// Update updates a task comment
func (s *TaskCommentService) Update(comment *TaskComment) error {
	return s.db.QueryRow(commentUpdateQuery, comment.ID, comment.CommentText).Scan(&comment.EditedAt)
}

// Delete deletes a task comment
func (s *TaskCommentService) Delete(id uuid.UUID) error {
	result, err := s.db.Exec(commentDeleteQuery, id)
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

// CanUserEdit checks if a user can edit a comment (author and within 15 minutes)
func (s *TaskCommentService) CanUserEdit(commentID, userID uuid.UUID) (bool, error) {
	var canEdit bool
	err := s.db.QueryRow(commentCanUserEditQuery, commentID, userID).Scan(&canEdit)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return canEdit, nil
}
