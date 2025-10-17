package models

// Query constants for TaskCommentService
const (
	commentCreateQuery = `
		INSERT INTO task_comments (id, task_id, user_id, comment_text)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`

	commentGetByTaskQuery = `
		SELECT tc.id, tc.task_id, tc.user_id, tc.comment_text, tc.created_at, tc.edited_at,
		       u.email as user_email,
		       COALESCE(v.name, a.name, u.email) as user_name
		FROM task_comments tc
		JOIN users u ON tc.user_id = u.id
		LEFT JOIN volunteers v ON u.id = v.user_id
		LEFT JOIN admins a ON u.id = a.user_id
		WHERE tc.task_id = $1
		ORDER BY tc.created_at ASC`

	commentGetByIDQuery = `
		SELECT id, task_id, user_id, comment_text, created_at, edited_at
		FROM task_comments WHERE id = $1`

	commentUpdateQuery = `
		UPDATE task_comments 
		SET comment_text = $2, edited_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING edited_at`

	commentDeleteQuery = `DELETE FROM task_comments WHERE id = $1`

	commentCanUserEditQuery = `
		SELECT 
			CASE 
				WHEN user_id = $2 AND created_at > (CURRENT_TIMESTAMP - INTERVAL '15 minutes') 
				THEN true 
				ELSE false 
			END
		FROM task_comments
		WHERE id = $1`
)
