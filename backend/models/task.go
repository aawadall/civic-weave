package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusTodo              TaskStatus = "todo"
	TaskStatusInProgress        TaskStatus = "in_progress"
	TaskStatusDone              TaskStatus = "done"
	TaskStatusBlocked           TaskStatus = "blocked"
	TaskStatusTakeoverRequested TaskStatus = "takeover_requested"
)

// TaskPriority represents the priority level of a task
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

// ProjectTask represents a task in a project
type ProjectTask struct {
	ID                  uuid.UUID    `json:"id" db:"id"`
	ProjectID           uuid.UUID    `json:"project_id" db:"project_id"`
	Title               string       `json:"title" db:"title"`
	Description         string       `json:"description" db:"description"`
	AssigneeID          *uuid.UUID   `json:"assignee_id" db:"assignee_id"`
	AssigneeName        *string      `json:"assignee_name,omitempty" db:"assignee_name"`
	AssigneeEmail       *string      `json:"assignee_email,omitempty" db:"assignee_email"`
	CreatedByID         uuid.UUID    `json:"created_by_id" db:"created_by_id"`
	Status              TaskStatus   `json:"status" db:"status"`
	Priority            TaskPriority `json:"priority" db:"priority"`
	DueDate             *time.Time   `json:"due_date" db:"due_date"`
	Labels              []string     `json:"labels" db:"labels"`
	ProjectTitle        string       `json:"project_title,omitempty" db:"project_title"`
	ProjectStatus       string       `json:"project_status,omitempty" db:"project_status"`
	StartedAt           *time.Time   `json:"started_at,omitempty" db:"started_at"`
	BlockedAt           *time.Time   `json:"blocked_at,omitempty" db:"blocked_at"`
	BlockedReason       *string      `json:"blocked_reason,omitempty" db:"blocked_reason"`
	CompletedAt         *time.Time   `json:"completed_at,omitempty" db:"completed_at"`
	CompletionNote      *string      `json:"completion_note,omitempty" db:"completion_note"`
	TakeoverRequestedAt *time.Time   `json:"takeover_requested_at,omitempty" db:"takeover_requested_at"`
	TakeoverReason      *string      `json:"takeover_reason,omitempty" db:"takeover_reason"`
	LastStatusChangedBy *uuid.UUID   `json:"last_status_changed_by,omitempty" db:"last_status_changed_by"`
	CreatedAt           time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at" db:"updated_at"`
}

// TaskUpdate represents a progress update on a task
type TaskUpdate struct {
	ID          uuid.UUID `json:"id" db:"id"`
	TaskID      uuid.UUID `json:"task_id" db:"task_id"`
	VolunteerID uuid.UUID `json:"volunteer_id" db:"volunteer_id"`
	UpdateText  string    `json:"update_text" db:"update_text"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// TaskWithUpdates includes the task and its updates
type TaskWithUpdates struct {
	ProjectTask
	Updates []TaskUpdate `json:"updates,omitempty"`
}

// TaskService handles task operations
type TaskService struct {
	db *sql.DB
}

// GetDB returns the database connection (for use by other services)
func (s *TaskService) GetDB() *sql.DB {
	return s.db
}

// NewTaskService creates a new task service
func NewTaskService(db *sql.DB) *TaskService {
	return &TaskService{db: db}
}

// Create creates a new task
func (s *TaskService) Create(task *ProjectTask) error {
	task.ID = uuid.New()
	labelsJSON, err := ToJSONArray(task.Labels)
	if err != nil {
		return err
	}

	return s.db.QueryRow(taskCreateQuery, task.ID, task.ProjectID, task.Title, task.Description,
		task.AssigneeID, task.CreatedByID, task.Status, task.Priority, task.DueDate, labelsJSON).
		Scan(&task.CreatedAt, &task.UpdatedAt)
}

// GetByID retrieves a task by ID
func (s *TaskService) GetByID(id uuid.UUID) (*ProjectTask, error) {
	task := &ProjectTask{}
	var labelsJSON string

	err := s.db.QueryRow(taskGetByIDQuery, id).Scan(
		&task.ID, &task.ProjectID, &task.Title, &task.Description, &task.AssigneeID,
		&task.CreatedByID, &task.Status, &task.Priority, &task.DueDate, &labelsJSON,
		&task.CreatedAt, &task.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse labels JSON
	if err := ParseJSONArray(labelsJSON, &task.Labels); err != nil {
		return nil, err
	}

	return task, nil
}

// ListByProject retrieves all tasks for a project
// If assigneeID is provided and is not the project owner, only returns tasks assigned to that volunteer
func (s *TaskService) ListByProject(projectID uuid.UUID, assigneeID *uuid.UUID, isProjectOwner bool) ([]ProjectTask, error) {
	var query string
	var args []interface{}

	if isProjectOwner || assigneeID == nil {
		// Project owner or admin sees all tasks
		query = taskListByProjectOwnerQuery
		args = []interface{}{projectID}
	} else {
		// Regular team member only sees their assigned tasks
		query = taskListByProjectMemberQuery
		args = []interface{}{projectID, assigneeID}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []ProjectTask
	for rows.Next() {
		var task ProjectTask
		var labelsJSON string
		err := rows.Scan(
			&task.ID, &task.ProjectID, &task.Title, &task.Description, &task.AssigneeID,
			&task.CreatedByID, &task.Status, &task.Priority, &task.DueDate, &labelsJSON,
			&task.CreatedAt, &task.UpdatedAt, &task.AssigneeName, &task.AssigneeEmail,
			&task.ProjectTitle, &task.ProjectStatus, &task.StartedAt, &task.BlockedAt,
			&task.BlockedReason, &task.CompletedAt, &task.CompletionNote,
			&task.TakeoverRequestedAt, &task.TakeoverReason, &task.LastStatusChangedBy,
		)
		if err != nil {
			return nil, err
		}

		// Parse labels JSON
		if err := ParseJSONArray(labelsJSON, &task.Labels); err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// ListUnassignedByProject retrieves unassigned tasks for self-assignment
func (s *TaskService) ListUnassignedByProject(projectID uuid.UUID) ([]ProjectTask, error) {
	rows, err := s.db.Query(taskListUnassignedByProjectQuery, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []ProjectTask
	for rows.Next() {
		var task ProjectTask
		var labelsJSON string
		err := rows.Scan(
			&task.ID, &task.ProjectID, &task.Title, &task.Description, &task.AssigneeID,
			&task.CreatedByID, &task.Status, &task.Priority, &task.DueDate, &labelsJSON,
			&task.CreatedAt, &task.UpdatedAt, &task.AssigneeName, &task.AssigneeEmail,
			&task.ProjectTitle, &task.ProjectStatus,
		)
		if err != nil {
			return nil, err
		}

		// Parse labels JSON
		if err := ParseJSONArray(labelsJSON, &task.Labels); err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// ListByAssignee retrieves tasks assigned to a specific volunteer
func (s *TaskService) ListByAssignee(assigneeID uuid.UUID) ([]ProjectTask, error) {
	rows, err := s.db.Query(taskListByAssigneeQuery, assigneeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []ProjectTask
	for rows.Next() {
		var task ProjectTask
		var labelsJSON string
		err := rows.Scan(
			&task.ID, &task.ProjectID, &task.Title, &task.Description, &task.AssigneeID,
			&task.CreatedByID, &task.Status, &task.Priority, &task.DueDate, &labelsJSON,
			&task.CreatedAt, &task.UpdatedAt, &task.AssigneeName, &task.AssigneeEmail,
			&task.ProjectTitle, &task.ProjectStatus, &task.StartedAt, &task.BlockedAt,
			&task.BlockedReason, &task.CompletedAt, &task.CompletionNote,
			&task.TakeoverRequestedAt, &task.TakeoverReason, &task.LastStatusChangedBy,
		)
		if err != nil {
			return nil, err
		}

		// Parse labels JSON
		if err := ParseJSONArray(labelsJSON, &task.Labels); err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// Update updates a task
func (s *TaskService) Update(task *ProjectTask) error {
	labelsJSON, err := ToJSONArray(task.Labels)
	if err != nil {
		return err
	}

	return s.db.QueryRow(taskUpdateQuery, task.ID, task.Title, task.Description, task.AssigneeID,
		task.Status, task.Priority, task.DueDate, labelsJSON).Scan(&task.UpdatedAt)
}

// UpdateStatus updates only the status of a task with activity logging
func (s *TaskService) UpdateStatus(taskID uuid.UUID, status TaskStatus, actorUserID uuid.UUID) error {
	// Get current task to determine previous status
	task, err := s.GetByID(taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return sql.ErrNoRows
	}

	// Update status
	result, err := s.db.Exec(taskUpdateStatusQuery, taskID, status, &actorUserID)
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

	// Log activity if status changed
	if task.Status != status {
		return s.insertActivityLog(taskID, actorUserID, task.Status, status, map[string]interface{}{})
	}

	return nil
}

// AssignToVolunteer assigns a task to a volunteer
func (s *TaskService) AssignToVolunteer(taskID, volunteerID uuid.UUID) error {
	result, err := s.db.Exec(taskAssignToVolunteerQuery, taskID, volunteerID)
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

// Delete deletes a task
func (s *TaskService) Delete(id uuid.UUID) error {
	result, err := s.db.Exec(taskDeleteQuery, id)
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

// AddUpdate adds a progress update to a task
func (s *TaskService) AddUpdate(update *TaskUpdate) error {
	update.ID = uuid.New()
	return s.db.QueryRow(taskAddUpdateQuery, update.ID, update.TaskID, update.VolunteerID, update.UpdateText).
		Scan(&update.CreatedAt)
}

// GetTaskUpdates retrieves all updates for a task
func (s *TaskService) GetTaskUpdates(taskID uuid.UUID) ([]TaskUpdate, error) {
	rows, err := s.db.Query(taskGetUpdatesQuery, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var updates []TaskUpdate
	for rows.Next() {
		var update TaskUpdate
		err := rows.Scan(&update.ID, &update.TaskID, &update.VolunteerID, &update.UpdateText, &update.CreatedAt)
		if err != nil {
			return nil, err
		}
		updates = append(updates, update)
	}

	return updates, rows.Err()
}

// GetTaskWithUpdates retrieves a task with all its updates
func (s *TaskService) GetTaskWithUpdates(taskID uuid.UUID) (*TaskWithUpdates, error) {
	task, err := s.GetByID(taskID)
	if err != nil || task == nil {
		return nil, err
	}

	updates, err := s.GetTaskUpdates(taskID)
	if err != nil {
		return nil, err
	}

	return &TaskWithUpdates{
		ProjectTask: *task,
		Updates:     updates,
	}, nil
}

// MarkAsBlocked marks a task as blocked
func (s *TaskService) MarkAsBlocked(taskID uuid.UUID, reason string, actorUserID uuid.UUID) error {
	// Get current task to determine previous status
	task, err := s.GetByID(taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return sql.ErrNoRows
	}

	// Update timeline fields and status
	now := time.Now()
	_, err = s.db.Exec(taskUpdateTimelineQuery, taskID, nil, &now, &reason, nil, nil, nil, nil, TaskStatusBlocked, &actorUserID)
	if err != nil {
		return err
	}

	// Log activity
	return s.insertActivityLog(taskID, actorUserID, task.Status, TaskStatusBlocked, map[string]interface{}{
		"reason": reason,
	})
}

// RequestTakeover marks a task as requesting takeover
func (s *TaskService) RequestTakeover(taskID uuid.UUID, reason string, actorUserID uuid.UUID) error {
	// Get current task to determine previous status
	task, err := s.GetByID(taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return sql.ErrNoRows
	}

	// Update timeline fields and status
	now := time.Now()
	_, err = s.db.Exec(taskUpdateTimelineQuery, taskID, nil, nil, nil, nil, nil, &now, &reason, TaskStatusTakeoverRequested, &actorUserID)
	if err != nil {
		return err
	}

	// Log activity
	return s.insertActivityLog(taskID, actorUserID, task.Status, TaskStatusTakeoverRequested, map[string]interface{}{
		"reason": reason,
	})
}

// MarkAsDone marks a task as done
func (s *TaskService) MarkAsDone(taskID uuid.UUID, completionNote string, actorUserID uuid.UUID) error {
	// Get current task to determine previous status
	task, err := s.GetByID(taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return sql.ErrNoRows
	}

	// Update timeline fields and status
	now := time.Now()
	_, err = s.db.Exec(taskUpdateTimelineQuery, taskID, nil, nil, nil, &now, &completionNote, nil, nil, TaskStatusDone, &actorUserID)
	if err != nil {
		return err
	}

	// Log activity
	return s.insertActivityLog(taskID, actorUserID, task.Status, TaskStatusDone, map[string]interface{}{
		"completion_note": completionNote,
	})
}

// StartTask marks a task as started (in_progress)
func (s *TaskService) StartTask(taskID uuid.UUID, actorUserID uuid.UUID) error {
	// Get current task to determine previous status
	task, err := s.GetByID(taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return sql.ErrNoRows
	}

	// Update timeline fields and status
	now := time.Now()
	_, err = s.db.Exec(taskUpdateTimelineQuery, taskID, &now, nil, nil, nil, nil, nil, nil, TaskStatusInProgress, &actorUserID)
	if err != nil {
		return err
	}

	// Log activity
	return s.insertActivityLog(taskID, actorUserID, task.Status, TaskStatusInProgress, map[string]interface{}{
		"started_at": now,
	})
}

// insertActivityLog logs a task status change to the activity log
func (s *TaskService) insertActivityLog(taskID, actorUserID uuid.UUID, fromStatus, toStatus TaskStatus, context map[string]interface{}) error {
	// Look up volunteer ID from user ID
	var volunteerID *uuid.UUID
	err := s.db.QueryRow("SELECT id FROM volunteers WHERE user_id = $1", actorUserID).Scan(&volunteerID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Prepare context JSON
	contextJSON, err := ToJSON(context)
	if err != nil {
		return err
	}

	// Insert activity log entry
	_, err = s.db.Exec(taskInsertActivityLogQuery, uuid.New(), taskID, actorUserID, volunteerID, fromStatus, toStatus, contextJSON)
	return err
}
