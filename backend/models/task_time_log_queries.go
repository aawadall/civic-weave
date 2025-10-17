package models

// Query constants for TaskTimeLogService
const (
	timeLogCreateQuery = `
		INSERT INTO task_time_logs (id, task_id, volunteer_id, hours, log_date, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`

	timeLogGetByTaskQuery = `
		SELECT ttl.id, ttl.task_id, ttl.volunteer_id, ttl.hours, ttl.log_date, 
		       ttl.description, ttl.created_at, v.name as volunteer_name
		FROM task_time_logs ttl
		JOIN volunteers v ON ttl.volunteer_id = v.id
		WHERE ttl.task_id = $1
		ORDER BY ttl.log_date DESC, ttl.created_at DESC`

	timeLogGetByVolunteerQuery = `
		SELECT id, task_id, volunteer_id, hours, log_date, description, created_at
		FROM task_time_logs
		WHERE volunteer_id = $1
		ORDER BY log_date DESC, created_at DESC`

	timeLogGetByProjectQuery = `
		SELECT ttl.id, ttl.task_id, ttl.volunteer_id, ttl.hours, ttl.log_date, 
		       ttl.description, ttl.created_at, v.name as volunteer_name
		FROM task_time_logs ttl
		JOIN volunteers v ON ttl.volunteer_id = v.id
		JOIN project_tasks pt ON ttl.task_id = pt.id
		WHERE pt.project_id = $1
		ORDER BY ttl.log_date DESC, ttl.created_at DESC`

	timeLogGetTotalByTaskQuery = `
		SELECT COALESCE(SUM(hours), 0) as total_hours
		FROM task_time_logs
		WHERE task_id = $1`

	timeLogGetTotalByVolunteerQuery = `
		SELECT COALESCE(SUM(ttl.hours), 0) as total_hours
		FROM task_time_logs ttl
		JOIN project_tasks pt ON ttl.task_id = pt.id
		WHERE ttl.volunteer_id = $1 AND pt.project_id = $2`

	timeLogGetTotalByProjectQuery = `
		SELECT COALESCE(SUM(ttl.hours), 0) as total_hours
		FROM task_time_logs ttl
		JOIN project_tasks pt ON ttl.task_id = pt.id
		WHERE pt.project_id = $1`

	timeLogGetSummaryByTaskQuery = `
		SELECT COALESCE(SUM(hours), 0) as total_hours, COUNT(*) as log_count
		FROM task_time_logs
		WHERE task_id = $1`

	timeLogGetSummaryByVolunteerQuery = `
		SELECT COALESCE(SUM(ttl.hours), 0) as total_hours, COUNT(*) as log_count
		FROM task_time_logs ttl
		JOIN project_tasks pt ON ttl.task_id = pt.id
		WHERE ttl.volunteer_id = $1 AND pt.project_id = $2`

	timeLogGetSummaryByProjectQuery = `
		SELECT COALESCE(SUM(ttl.hours), 0) as total_hours, COUNT(*) as log_count
		FROM task_time_logs ttl
		JOIN project_tasks pt ON ttl.task_id = pt.id
		WHERE pt.project_id = $1`

	timeLogDeleteQuery = `DELETE FROM task_time_logs WHERE id = $1`
)
