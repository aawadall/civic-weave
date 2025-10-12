package models

// Query constants for TaskService
const (
	taskCreateQuery = `
		INSERT INTO project_tasks (id, project_id, title, description, assignee_id, 
		                          created_by_id, status, priority, due_date, labels)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at`

	taskGetByIDQuery = `
		SELECT id, project_id, title, description, assignee_id, created_by_id, 
		       status, priority, due_date, labels, created_at, updated_at
		FROM project_tasks WHERE id = $1`

	taskListByProjectOwnerQuery = `
		SELECT id, project_id, title, description, assignee_id, created_by_id, 
		       status, priority, due_date, labels, created_at, updated_at
		FROM project_tasks 
		WHERE project_id = $1
		ORDER BY 
			CASE priority 
				WHEN 'high' THEN 1 
				WHEN 'medium' THEN 2 
				WHEN 'low' THEN 3 
			END,
			created_at DESC`

	taskListByProjectMemberQuery = `
		SELECT id, project_id, title, description, assignee_id, created_by_id, 
		       status, priority, due_date, labels, created_at, updated_at
		FROM project_tasks 
		WHERE project_id = $1 AND (assignee_id = $2 OR assignee_id IS NULL)
		ORDER BY 
			CASE priority 
				WHEN 'high' THEN 1 
				WHEN 'medium' THEN 2 
				WHEN 'low' THEN 3 
			END,
			created_at DESC`

	taskListUnassignedByProjectQuery = `
		SELECT id, project_id, title, description, assignee_id, created_by_id, 
		       status, priority, due_date, labels, created_at, updated_at
		FROM project_tasks 
		WHERE project_id = $1 AND assignee_id IS NULL AND status != 'done'
		ORDER BY 
			CASE priority 
				WHEN 'high' THEN 1 
				WHEN 'medium' THEN 2 
				WHEN 'low' THEN 3 
			END,
			created_at DESC`

	taskListByAssigneeQuery = `
		SELECT id, project_id, title, description, assignee_id, created_by_id, 
		       status, priority, due_date, labels, created_at, updated_at
		FROM project_tasks 
		WHERE assignee_id = $1
		ORDER BY 
			CASE status 
				WHEN 'in_progress' THEN 1 
				WHEN 'todo' THEN 2 
				WHEN 'done' THEN 3 
			END,
			CASE priority 
				WHEN 'high' THEN 1 
				WHEN 'medium' THEN 2 
				WHEN 'low' THEN 3 
			END,
			due_date ASC NULLS LAST`

	taskUpdateQuery = `
		UPDATE project_tasks 
		SET title = $2, description = $3, assignee_id = $4, status = $5, 
		    priority = $6, due_date = $7, labels = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	taskUpdateStatusQuery = `
		UPDATE project_tasks 
		SET status = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	taskAssignToVolunteerQuery = `
		UPDATE project_tasks 
		SET assignee_id = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	taskDeleteQuery = `DELETE FROM project_tasks WHERE id = $1`

	taskAddUpdateQuery = `
		INSERT INTO task_updates (id, task_id, volunteer_id, update_text)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`

	taskGetUpdatesQuery = `
		SELECT id, task_id, volunteer_id, update_text, created_at
		FROM task_updates
		WHERE task_id = $1
		ORDER BY created_at DESC`
)

