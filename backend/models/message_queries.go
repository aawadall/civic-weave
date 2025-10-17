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

	messageCreateUniversalQuery = `
		INSERT INTO project_messages (id, project_id, sender_id, recipient_user_id, recipient_team_id, 
		                              subject, message_text, task_id, message_type, message_scope)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at`

	messageGetInboxQuery = `
		SELECT 
			pm.id, pm.project_id, pm.sender_id, pm.recipient_user_id, pm.recipient_team_id,
			pm.subject, pm.message_text, pm.task_id, pm.message_type, pm.message_scope,
			pm.created_at, pm.edited_at, pm.deleted_at,
			COALESCE(sender_v.name, sender_a.name, sender_u.email) as sender_name,
			sender_u.email as sender_email,
			COALESCE(recipient_v.name, recipient_a.name, recipient_u.email) as recipient_name,
			recipient_u.email as recipient_email,
			p.title as project_title,
			CASE WHEN mr.user_id IS NOT NULL THEN true ELSE false END as is_read
		FROM project_messages pm
		JOIN users sender_u ON pm.sender_id = sender_u.id
		LEFT JOIN volunteers sender_v ON sender_u.id = sender_v.user_id
		LEFT JOIN admins sender_a ON sender_u.id = sender_a.user_id
		LEFT JOIN users recipient_u ON pm.recipient_user_id = recipient_u.id
		LEFT JOIN volunteers recipient_v ON recipient_u.id = recipient_v.user_id
		LEFT JOIN admins recipient_a ON recipient_u.id = recipient_a.user_id
		LEFT JOIN projects p ON pm.project_id = p.id
		LEFT JOIN message_reads mr ON pm.id = mr.message_id AND mr.user_id = $1
		WHERE pm.deleted_at IS NULL
		AND (
			pm.recipient_user_id = $1 OR
			(pm.recipient_team_id IS NOT NULL AND EXISTS (
				SELECT 1 FROM project_team_members ptm 
				JOIN volunteers v ON ptm.volunteer_id = v.id 
				WHERE ptm.project_id = pm.recipient_team_id AND v.user_id = $1 AND ptm.status = 'active'
			)) OR
			(pm.project_id IS NOT NULL AND EXISTS (
				SELECT 1 FROM project_team_members ptm 
				JOIN volunteers v ON ptm.volunteer_id = v.id 
				WHERE ptm.project_id = pm.project_id AND v.user_id = $1 AND ptm.status = 'active'
			))
		)
		ORDER BY pm.created_at DESC
		LIMIT $2 OFFSET $3`

	messageGetSentQuery = `
		SELECT 
			pm.id, pm.project_id, pm.sender_id, pm.recipient_user_id, pm.recipient_team_id,
			pm.subject, pm.message_text, pm.task_id, pm.message_type, pm.message_scope,
			pm.created_at, pm.edited_at, pm.deleted_at,
			COALESCE(sender_v.name, sender_a.name, sender_u.email) as sender_name,
			sender_u.email as sender_email,
			COALESCE(recipient_v.name, recipient_a.name, recipient_u.email) as recipient_name,
			recipient_u.email as recipient_email,
			p.title as project_title,
			true as is_read
		FROM project_messages pm
		JOIN users sender_u ON pm.sender_id = sender_u.id
		LEFT JOIN volunteers sender_v ON sender_u.id = sender_v.user_id
		LEFT JOIN admins sender_a ON sender_u.id = sender_a.user_id
		LEFT JOIN users recipient_u ON pm.recipient_user_id = recipient_u.id
		LEFT JOIN volunteers recipient_v ON recipient_u.id = recipient_v.user_id
		LEFT JOIN admins recipient_a ON recipient_u.id = recipient_a.user_id
		LEFT JOIN projects p ON pm.project_id = p.id
		WHERE pm.deleted_at IS NULL
		AND pm.sender_id = $1
		ORDER BY pm.created_at DESC
		LIMIT $2 OFFSET $3`

	messageGetConversationsQuery = `
		WITH conversation_summary AS (
			SELECT 
				CASE 
					WHEN pm.recipient_user_id IS NOT NULL THEN pm.recipient_user_id
					WHEN pm.recipient_team_id IS NOT NULL THEN pm.recipient_team_id
					ELSE pm.project_id
				END as conversation_id,
				CASE 
					WHEN pm.recipient_user_id IS NOT NULL THEN 'user_to_user'
					WHEN pm.recipient_team_id IS NOT NULL THEN 'user_to_team'
					ELSE 'project'
				END as conversation_type,
				CASE 
					WHEN pm.recipient_user_id IS NOT NULL THEN COALESCE(recipient_v.name, recipient_a.name, recipient_u.email)
					WHEN pm.recipient_team_id IS NOT NULL THEN p.title
					ELSE p.title
				END as conversation_title,
				COUNT(*) as unread_count,
				MAX(pm.created_at) as last_activity,
				COUNT(DISTINCT pm.sender_id) as participant_count,
				MAX(pm.id) as last_message_id,
				MAX(pm.message_text) as last_message_text,
				MAX(pm.created_at) as last_message_created_at,
				MAX(COALESCE(sender_v.name, sender_a.name, sender_u.email)) as last_sender_name,
				MAX(sender_u.email) as last_sender_email
			FROM project_messages pm
			JOIN users sender_u ON pm.sender_id = sender_u.id
			LEFT JOIN volunteers sender_v ON sender_u.id = sender_v.user_id
			LEFT JOIN admins sender_a ON sender_u.id = sender_a.user_id
			LEFT JOIN users recipient_u ON pm.recipient_user_id = recipient_u.id
			LEFT JOIN volunteers recipient_v ON recipient_u.id = recipient_v.user_id
			LEFT JOIN admins recipient_a ON recipient_u.id = recipient_a.user_id
			LEFT JOIN projects p ON (pm.project_id = p.id OR pm.recipient_team_id = p.id)
			LEFT JOIN message_reads mr ON pm.id = mr.message_id AND mr.user_id = $1
			WHERE pm.deleted_at IS NULL
			AND (
				pm.recipient_user_id = $1 OR
				(pm.recipient_team_id IS NOT NULL AND EXISTS (
					SELECT 1 FROM project_team_members ptm 
					JOIN volunteers v ON ptm.volunteer_id = v.id 
					WHERE ptm.project_id = pm.recipient_team_id AND v.user_id = $1 AND ptm.status = 'active'
				)) OR
				(pm.project_id IS NOT NULL AND EXISTS (
					SELECT 1 FROM project_team_members ptm 
					JOIN volunteers v ON ptm.volunteer_id = v.id 
					WHERE ptm.project_id = pm.project_id AND v.user_id = $1 AND ptm.status = 'active'
				)) OR
				pm.sender_id = $1
			)
			GROUP BY conversation_id, conversation_type, conversation_title
		)
		SELECT 
			conversation_id::uuid as id,
			conversation_type as type,
			conversation_title as title,
			unread_count,
			participant_count,
			last_activity as updated_at,
			last_message_id,
			last_message_text,
			last_message_created_at,
			last_sender_name,
			last_sender_email
		FROM conversation_summary
		ORDER BY last_activity DESC
		LIMIT $2 OFFSET $3`

	messageGetConversationQuery = `
		SELECT 
			pm.id, pm.project_id, pm.sender_id, pm.recipient_user_id, pm.recipient_team_id,
			pm.subject, pm.message_text, pm.task_id, pm.message_type, pm.message_scope,
			pm.created_at, pm.edited_at, pm.deleted_at,
			COALESCE(sender_v.name, sender_a.name, sender_u.email) as sender_name,
			sender_u.email as sender_email,
			COALESCE(recipient_v.name, recipient_a.name, recipient_u.email) as recipient_name,
			recipient_u.email as recipient_email,
			p.title as project_title,
			CASE WHEN mr.user_id IS NOT NULL THEN true ELSE false END as is_read
		FROM project_messages pm
		JOIN users sender_u ON pm.sender_id = sender_u.id
		LEFT JOIN volunteers sender_v ON sender_u.id = sender_v.user_id
		LEFT JOIN admins sender_a ON sender_u.id = sender_a.user_id
		LEFT JOIN users recipient_u ON pm.recipient_user_id = recipient_u.id
		LEFT JOIN volunteers recipient_v ON recipient_u.id = recipient_v.user_id
		LEFT JOIN admins recipient_a ON recipient_u.id = recipient_a.user_id
		LEFT JOIN projects p ON pm.project_id = p.id
		LEFT JOIN message_reads mr ON pm.id = mr.message_id AND mr.user_id = $2
		WHERE pm.deleted_at IS NULL
		AND (
			(pm.recipient_user_id = $1 AND pm.sender_id = $2) OR
			(pm.recipient_user_id = $2 AND pm.sender_id = $1) OR
			(pm.recipient_team_id = $1 AND EXISTS (
				SELECT 1 FROM project_team_members ptm 
				JOIN volunteers v ON ptm.volunteer_id = v.id 
				WHERE ptm.project_id = pm.recipient_team_id AND v.user_id = $2 AND ptm.status = 'active'
			)) OR
			(pm.project_id = $1 AND EXISTS (
				SELECT 1 FROM project_team_members ptm 
				JOIN volunteers v ON ptm.volunteer_id = v.id 
				WHERE ptm.project_id = pm.project_id AND v.user_id = $2 AND ptm.status = 'active'
			))
		)
		ORDER BY pm.created_at ASC
		LIMIT $3 OFFSET $4`

	messageGetUniversalUnreadCountQuery = `
		SELECT 
			COUNT(*) as total,
			0 as direct_messages,
			0 as team_messages,
			0 as project_messages
		FROM project_messages pm
		LEFT JOIN message_reads mr ON pm.id = mr.message_id AND mr.user_id = $1
		WHERE pm.deleted_at IS NULL
		  AND mr.user_id IS NULL
		  AND pm.sender_id != $1
		  AND EXISTS (
			SELECT 1 FROM project_team_members ptm 
			JOIN volunteers v ON ptm.volunteer_id = v.id 
			WHERE ptm.project_id = pm.project_id AND v.user_id = $1 AND ptm.status = 'active'
		)`

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

	messageSearchUsersQuery = `
		SELECT DISTINCT u.id, COALESCE(v.name, a.name, u.email) as name, u.email
		FROM users u
		LEFT JOIN volunteers v ON u.id = v.user_id
		LEFT JOIN admins a ON u.id = a.user_id
		WHERE (
			LOWER(COALESCE(v.name, a.name, u.email)) LIKE LOWER('%' || $1 || '%') OR
			LOWER(u.email) LIKE LOWER('%' || $1 || '%')
		)
		ORDER BY name
		LIMIT $2`

	messageSearchUserProjectsQuery = `
		SELECT DISTINCT p.id, p.title
		FROM projects p
		JOIN project_team_members ptm ON p.id = ptm.project_id
		JOIN volunteers v ON ptm.volunteer_id = v.id
		WHERE v.user_id = $1 
		  AND ptm.status = 'active'
		  AND LOWER(p.title) LIKE LOWER('%' || $2 || '%')
		ORDER BY p.title
		LIMIT $3`
)
