package models

// Query constants for MessageService
const (
	messageCreateQuery = `
		INSERT INTO project_messages (id, project_id, sender_id, message_text, task_id, message_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`

	messageGetByIDQuery = `
		SELECT id, project_id, sender_id, message_text, task_id, message_type, created_at, edited_at, deleted_at
		FROM project_messages WHERE id = $1`

	messageListByProjectQuery = `
		SELECT 
			pm.id, pm.project_id, pm.sender_id, pm.message_text, 
			pm.created_at, pm.edited_at, pm.deleted_at,
			u.email as sender_email,
			COALESCE(v.name, a.name, u.email) as sender_name,
			CASE WHEN mr.user_id IS NOT NULL THEN true ELSE false END as is_read
		FROM project_messages pm
		JOIN users u ON pm.sender_id = u.id
		LEFT JOIN volunteers v ON u.id = v.user_id
		LEFT JOIN admins a ON u.id = a.user_id
		LEFT JOIN message_reads mr ON pm.id = mr.message_id AND mr.user_id = $4
		WHERE pm.project_id = $1 AND pm.deleted_at IS NULL
		ORDER BY pm.created_at ASC
		LIMIT $2 OFFSET $3`

	messageListRecentByProjectQuery = `
		SELECT 
			pm.id, pm.project_id, pm.sender_id, pm.message_text, 
			pm.created_at, pm.edited_at, pm.deleted_at,
			u.email as sender_email,
			COALESCE(v.name, a.name, u.email) as sender_name,
			CASE WHEN mr.user_id IS NOT NULL THEN true ELSE false END as is_read
		FROM project_messages pm
		JOIN users u ON pm.sender_id = u.id
		LEFT JOIN volunteers v ON u.id = v.user_id
		LEFT JOIN admins a ON u.id = a.user_id
		LEFT JOIN message_reads mr ON pm.id = mr.message_id AND mr.user_id = $3
		WHERE pm.project_id = $1 AND pm.deleted_at IS NULL
		ORDER BY pm.created_at DESC
		LIMIT $2`

	messageUpdateQuery = `
		UPDATE project_messages 
		SET message_text = $2, edited_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING edited_at`

	messageSoftDeleteQuery = `
		UPDATE project_messages 
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL`

	messageMarkAsReadQuery = `
		INSERT INTO message_reads (user_id, message_id, read_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, message_id) DO NOTHING`

	messageMarkAllAsReadQuery = `
		INSERT INTO message_reads (user_id, message_id, read_at)
		SELECT $1, pm.id, CURRENT_TIMESTAMP
		FROM project_messages pm
		WHERE pm.project_id = $2 AND pm.deleted_at IS NULL
		ON CONFLICT (user_id, message_id) DO NOTHING`

	messageGetUnreadCountQuery = `
		SELECT COUNT(*)
		FROM project_messages pm
		LEFT JOIN message_reads mr ON pm.id = mr.message_id AND mr.user_id = $2
		WHERE pm.project_id = $1 
		  AND pm.deleted_at IS NULL
		  AND pm.sender_id != $2
		  AND mr.user_id IS NULL`

	messageGetUnreadCountsByUserQuery = `
		SELECT 
			pm.project_id,
			COUNT(*) as unread_count
		FROM project_messages pm
		JOIN project_team_members ptm ON pm.project_id = ptm.project_id
		LEFT JOIN message_reads mr ON pm.id = mr.message_id AND mr.user_id = $1
		WHERE ptm.volunteer_id = (
			SELECT id FROM volunteers WHERE user_id = $1
		)
		  AND ptm.status = 'active'
		  AND pm.deleted_at IS NULL
		  AND pm.sender_id != $1
		  AND mr.user_id IS NULL
		GROUP BY pm.project_id`

	messageGetMessagesAfterQuery = `
		SELECT 
			pm.id, pm.project_id, pm.sender_id, pm.message_text, 
			pm.created_at, pm.edited_at, pm.deleted_at,
			u.email as sender_email,
			COALESCE(v.name, a.name, u.email) as sender_name,
			CASE WHEN mr.user_id IS NOT NULL THEN true ELSE false END as is_read
		FROM project_messages pm
		JOIN users u ON pm.sender_id = u.id
		LEFT JOIN volunteers v ON u.id = v.user_id
		LEFT JOIN admins a ON u.id = a.user_id
		LEFT JOIN message_reads mr ON pm.id = mr.message_id AND mr.user_id = $3
		WHERE pm.project_id = $1 
		  AND pm.created_at > $2 
		  AND pm.deleted_at IS NULL
		ORDER BY pm.created_at ASC`

	messageCanUserEditQuery = `
		SELECT 
			CASE 
				WHEN sender_id = $2 AND created_at > (CURRENT_TIMESTAMP - INTERVAL '15 minutes') 
				THEN true 
				ELSE false 
			END
		FROM project_messages
		WHERE id = $1`
)
