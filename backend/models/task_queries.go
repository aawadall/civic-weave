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
		SELECT pt.id, pt.project_id, pt.title, pt.description, pt.assignee_id, pt.created_by_id, 
		       pt.status, pt.priority, pt.due_date, pt.labels, pt.created_at, pt.updated_at,
		       v.name as assignee_name, u.email as assignee_email,
		       p.title as project_title, p.project_status,
		       pt.started_at, pt.blocked_at, pt.blocked_reason, pt.completed_at, pt.completion_note,
		       pt.takeover_requested_at, pt.takeover_reason, pt.last_status_changed_by
		FROM project_tasks pt
		LEFT JOIN volunteers v ON pt.assignee_id = v.id
		LEFT JOIN users u ON v.user_id = u.id
		LEFT JOIN projects p ON pt.project_id = p.id
		WHERE pt.project_id = $1
		ORDER BY 
			CASE pt.priority 
				WHEN 'high' THEN 1 
				WHEN 'medium' THEN 2 
				WHEN 'low' THEN 3 
			END,
			pt.created_at DESC`

	taskListByProjectMemberQuery = `
		SELECT pt.id, pt.project_id, pt.title, pt.description, pt.assignee_id, pt.created_by_id, 
		       pt.status, pt.priority, pt.due_date, pt.labels, pt.created_at, pt.updated_at,
		       v.name as assignee_name, u.email as assignee_email,
		       p.title as project_title, p.project_status,
		       pt.started_at, pt.blocked_at, pt.blocked_reason, pt.completed_at, pt.completion_note,
		       pt.takeover_requested_at, pt.takeover_reason, pt.last_status_changed_by
		FROM project_tasks pt
		LEFT JOIN volunteers v ON pt.assignee_id = v.id
		LEFT JOIN users u ON v.user_id = u.id
		LEFT JOIN projects p ON pt.project_id = p.id
		WHERE pt.project_id = $1 AND (pt.assignee_id = $2 OR pt.assignee_id IS NULL)
		ORDER BY 
			CASE pt.priority 
				WHEN 'high' THEN 1 
				WHEN 'medium' THEN 2 
				WHEN 'low' THEN 3 
			END,
			pt.created_at DESC`

	taskListUnassignedByProjectQuery = `
		SELECT pt.id, pt.project_id, pt.title, pt.description, pt.assignee_id, pt.created_by_id, 
		       pt.status, pt.priority, pt.due_date, pt.labels, pt.created_at, pt.updated_at,
		       v.name as assignee_name, u.email as assignee_email,
		       p.title as project_title, p.project_status
		FROM project_tasks pt
		LEFT JOIN volunteers v ON pt.assignee_id = v.id
		LEFT JOIN users u ON v.user_id = u.id
		LEFT JOIN projects p ON pt.project_id = p.id
		WHERE pt.project_id = $1 AND pt.assignee_id IS NULL AND pt.status != 'done'
		ORDER BY 
			CASE pt.priority 
				WHEN 'high' THEN 1 
				WHEN 'medium' THEN 2 
				WHEN 'low' THEN 3 
			END,
			pt.created_at DESC`

	taskListByAssigneeQuery = `
		SELECT pt.id, pt.project_id, pt.title, pt.description, pt.assignee_id, pt.created_by_id, 
		       pt.status, pt.priority, pt.due_date, pt.labels, pt.created_at, pt.updated_at,
		       v.name as assignee_name, u.email as assignee_email,
		       p.title as project_title, p.project_status,
		       pt.started_at, pt.blocked_at, pt.blocked_reason, pt.completed_at, pt.completion_note,
		       pt.takeover_requested_at, pt.takeover_reason, pt.last_status_changed_by
		FROM project_tasks pt
		JOIN projects p ON pt.project_id = p.id
		JOIN volunteers v ON pt.assignee_id = v.id
		JOIN users u ON v.user_id = u.id
		WHERE pt.assignee_id = $1
		ORDER BY 
			CASE pt.status 
				WHEN 'in_progress' THEN 1 
				WHEN 'todo' THEN 2 
				WHEN 'blocked' THEN 3
				WHEN 'takeover_requested' THEN 4
				WHEN 'done' THEN 5 
			END,
			CASE pt.priority 
				WHEN 'high' THEN 1 
				WHEN 'medium' THEN 2 
				WHEN 'low' THEN 3 
			END,
			pt.due_date ASC NULLS LAST`

	taskUpdateQuery = `
		UPDATE project_tasks 
		SET title = $2, description = $3, assignee_id = $4, status = $5, 
		    priority = $6, due_date = $7, labels = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	taskUpdateStatusQuery = `
		UPDATE project_tasks 
		SET status = $2, last_status_changed_by = $3, updated_at = CURRENT_TIMESTAMP
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

	taskInsertActivityLogQuery = `
		INSERT INTO task_activity_log (id, task_id, actor_user_id, actor_volunteer_id, from_status, to_status, context)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	taskUpdateTimelineQuery = `
		UPDATE project_tasks 
		SET started_at = $2, blocked_at = $3, blocked_reason = $4, completed_at = $5, 
		    completion_note = $6, takeover_requested_at = $7, takeover_reason = $8, 
		    status = $9, last_status_changed_by = $10, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`
)
