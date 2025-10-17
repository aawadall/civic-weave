package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// TaskTimeLog represents a time log entry for a task
type TaskTimeLog struct {
	ID          uuid.UUID `json:"id" db:"id"`
	TaskID      uuid.UUID `json:"task_id" db:"task_id"`
	VolunteerID uuid.UUID `json:"volunteer_id" db:"volunteer_id"`
	Hours       float64   `json:"hours" db:"hours"`
	LogDate     time.Time `json:"log_date" db:"log_date"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// TimeLogWithVolunteer includes volunteer information
type TimeLogWithVolunteer struct {
	TaskTimeLog
	VolunteerName string `json:"volunteer_name"`
}

// TimeSummary represents aggregated time data
type TimeSummary struct {
	TotalHours float64 `json:"total_hours"`
	LogCount   int     `json:"log_count"`
}

// TaskTimeLogService handles time log operations
type TaskTimeLogService struct {
	db *sql.DB
}

// NewTaskTimeLogService creates a new time log service
func NewTaskTimeLogService(db *sql.DB) *TaskTimeLogService {
	return &TaskTimeLogService{db: db}
}

// Create creates a new time log entry
func (s *TaskTimeLogService) Create(timeLog *TaskTimeLog) error {
	timeLog.ID = uuid.New()
	return s.db.QueryRow(timeLogCreateQuery, timeLog.ID, timeLog.TaskID, timeLog.VolunteerID, 
		timeLog.Hours, timeLog.LogDate, timeLog.Description).Scan(&timeLog.CreatedAt)
}

// GetByTask retrieves all time logs for a specific task
func (s *TaskTimeLogService) GetByTask(taskID uuid.UUID) ([]TimeLogWithVolunteer, error) {
	rows, err := s.db.Query(timeLogGetByTaskQuery, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var timeLogs []TimeLogWithVolunteer
	for rows.Next() {
		var log TimeLogWithVolunteer
		err := rows.Scan(
			&log.ID, &log.TaskID, &log.VolunteerID, &log.Hours, 
			&log.LogDate, &log.Description, &log.CreatedAt, &log.VolunteerName,
		)
		if err != nil {
			return nil, err
		}
		timeLogs = append(timeLogs, log)
	}

	return timeLogs, rows.Err()
}

// GetByVolunteer retrieves all time logs for a specific volunteer
func (s *TaskTimeLogService) GetByVolunteer(volunteerID uuid.UUID) ([]TaskTimeLog, error) {
	rows, err := s.db.Query(timeLogGetByVolunteerQuery, volunteerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var timeLogs []TaskTimeLog
	for rows.Next() {
		var log TaskTimeLog
		err := rows.Scan(
			&log.ID, &log.TaskID, &log.VolunteerID, &log.Hours, 
			&log.LogDate, &log.Description, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		timeLogs = append(timeLogs, log)
	}

	return timeLogs, rows.Err()
}

// GetByProject retrieves all time logs for a specific project
func (s *TaskTimeLogService) GetByProject(projectID uuid.UUID) ([]TimeLogWithVolunteer, error) {
	rows, err := s.db.Query(timeLogGetByProjectQuery, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var timeLogs []TimeLogWithVolunteer
	for rows.Next() {
		var log TimeLogWithVolunteer
		err := rows.Scan(
			&log.ID, &log.TaskID, &log.VolunteerID, &log.Hours, 
			&log.LogDate, &log.Description, &log.CreatedAt, &log.VolunteerName,
		)
		if err != nil {
			return nil, err
		}
		timeLogs = append(timeLogs, log)
	}

	return timeLogs, rows.Err()
}

// GetTotalHoursByTask returns total hours logged for a specific task
func (s *TaskTimeLogService) GetTotalHoursByTask(taskID uuid.UUID) (float64, error) {
	var totalHours float64
	err := s.db.QueryRow(timeLogGetTotalByTaskQuery, taskID).Scan(&totalHours)
	if err != nil {
		return 0, err
	}
	return totalHours, nil
}

// GetTotalHoursByVolunteer returns total hours logged by a volunteer for a specific project
func (s *TaskTimeLogService) GetTotalHoursByVolunteer(volunteerID, projectID uuid.UUID) (float64, error) {
	var totalHours float64
	err := s.db.QueryRow(timeLogGetTotalByVolunteerQuery, volunteerID, projectID).Scan(&totalHours)
	if err != nil {
		return 0, err
	}
	return totalHours, nil
}

// GetTotalHoursByProject returns total hours logged for a specific project
func (s *TaskTimeLogService) GetTotalHoursByProject(projectID uuid.UUID) (float64, error) {
	var totalHours float64
	err := s.db.QueryRow(timeLogGetTotalByProjectQuery, projectID).Scan(&totalHours)
	if err != nil {
		return 0, err
	}
	return totalHours, nil
}

// GetTimeSummaryByTask returns time summary for a specific task
func (s *TaskTimeLogService) GetTimeSummaryByTask(taskID uuid.UUID) (*TimeSummary, error) {
	summary := &TimeSummary{}
	err := s.db.QueryRow(timeLogGetSummaryByTaskQuery, taskID).Scan(&summary.TotalHours, &summary.LogCount)
	if err != nil {
		return nil, err
	}
	return summary, nil
}

// GetTimeSummaryByVolunteer returns time summary for a volunteer in a project
func (s *TaskTimeLogService) GetTimeSummaryByVolunteer(volunteerID, projectID uuid.UUID) (*TimeSummary, error) {
	summary := &TimeSummary{}
	err := s.db.QueryRow(timeLogGetSummaryByVolunteerQuery, volunteerID, projectID).Scan(&summary.TotalHours, &summary.LogCount)
	if err != nil {
		return nil, err
	}
	return summary, nil
}

// GetTimeSummaryByProject returns time summary for a project
func (s *TaskTimeLogService) GetTimeSummaryByProject(projectID uuid.UUID) (*TimeSummary, error) {
	summary := &TimeSummary{}
	err := s.db.QueryRow(timeLogGetSummaryByProjectQuery, projectID).Scan(&summary.TotalHours, &summary.LogCount)
	if err != nil {
		return nil, err
	}
	return summary, nil
}

// Delete deletes a time log entry
func (s *TaskTimeLogService) Delete(id uuid.UUID) error {
	result, err := s.db.Exec(timeLogDeleteQuery, id)
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
