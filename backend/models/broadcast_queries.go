package models

// Query constants for BroadcastService
const (
	broadcastCreateQuery = `
		INSERT INTO broadcast_messages (id, title, content, author_id, target_audience, priority, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`

	broadcastGetByIDQuery = `
		SELECT id, title, content, author_id, target_audience, priority, expires_at, 
		       created_at, updated_at, deleted_at
		FROM broadcast_messages WHERE id = $1 AND deleted_at IS NULL`

	broadcastListQuery = `
		SELECT 
			bm.id, bm.title, bm.content, bm.author_id, bm.target_audience, bm.priority, 
			bm.expires_at, bm.created_at, bm.updated_at, bm.deleted_at,
			COALESCE(v.name, a.name, u.email) as author_name,
			u.email as author_email,
			CASE WHEN br.user_id IS NOT NULL THEN true ELSE false END as is_read
		FROM broadcast_messages bm
		JOIN users u ON bm.author_id = u.id
		LEFT JOIN volunteers v ON u.id = v.user_id
		LEFT JOIN admins a ON u.id = a.user_id
		LEFT JOIN broadcast_reads br ON bm.id = br.broadcast_id AND br.user_id = $1
		WHERE bm.deleted_at IS NULL
		AND (bm.expires_at IS NULL OR bm.expires_at > NOW())
		AND (
			bm.target_audience = 'all_users' OR
			(bm.target_audience = 'volunteers_only' AND $2 = 'volunteer') OR
			(bm.target_audience = 'admins_only' AND $2 = 'admin') OR
			(bm.target_audience = 'team_leads_only' AND $2 = 'team_lead')
		)
		ORDER BY 
			CASE bm.priority 
				WHEN 'urgent' THEN 1 
				WHEN 'high' THEN 2 
				WHEN 'normal' THEN 3 
				WHEN 'low' THEN 4 
			END,
			bm.created_at DESC
		LIMIT $3 OFFSET $4`

	broadcastListAllQuery = `
		SELECT 
			bm.id, bm.title, bm.content, bm.author_id, bm.target_audience, bm.priority, 
			bm.expires_at, bm.created_at, bm.updated_at, bm.deleted_at,
			COALESCE(v.name, a.name, u.email) as author_name,
			u.email as author_email,
			false as is_read
		FROM broadcast_messages bm
		JOIN users u ON bm.author_id = u.id
		LEFT JOIN volunteers v ON u.id = v.user_id
		LEFT JOIN admins a ON u.id = a.user_id
		WHERE bm.deleted_at IS NULL
		ORDER BY bm.created_at DESC
		LIMIT $1 OFFSET $2`

	broadcastUpdateQuery = `
		UPDATE broadcast_messages 
		SET title = $2, content = $3, target_audience = $4, priority = $5, 
		    expires_at = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at`

	broadcastSoftDeleteQuery = `
		UPDATE broadcast_messages 
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL`

	broadcastMarkAsReadQuery = `
		INSERT INTO broadcast_reads (user_id, broadcast_id, read_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, broadcast_id) DO NOTHING`

	broadcastGetUnreadCountQuery = `
		SELECT COUNT(*)
		FROM broadcast_messages bm
		LEFT JOIN broadcast_reads br ON bm.id = br.broadcast_id AND br.user_id = $1
		WHERE bm.deleted_at IS NULL
		AND (bm.expires_at IS NULL OR bm.expires_at > NOW())
		AND br.user_id IS NULL
		AND (
			bm.target_audience = 'all_users' OR
			(bm.target_audience = 'volunteers_only' AND 'volunteer' = ANY($2)) OR
			(bm.target_audience = 'admins_only' AND 'admin' = ANY($2)) OR
			(bm.target_audience = 'team_leads_only' AND 'team_lead' = ANY($2))
		)`

	broadcastGetStatsQuery = `
		SELECT 
			COUNT(*) as total_broadcasts,
			COUNT(CASE WHEN br.user_id IS NULL THEN 1 END) as unread_broadcasts,
			COUNT(CASE WHEN bm.priority = 'high' THEN 1 END) as high_priority_count,
			COUNT(CASE WHEN bm.priority = 'urgent' THEN 1 END) as urgent_count
		FROM broadcast_messages bm
		LEFT JOIN broadcast_reads br ON bm.id = br.broadcast_id AND br.user_id = $1
		WHERE bm.deleted_at IS NULL
		AND (bm.expires_at IS NULL OR bm.expires_at > NOW())
		AND (
			bm.target_audience = 'all_users' OR
			(bm.target_audience = 'volunteers_only' AND 'volunteer' = ANY($2)) OR
			(bm.target_audience = 'admins_only' AND 'admin' = ANY($2)) OR
			(bm.target_audience = 'team_leads_only' AND 'team_lead' = ANY($2))
		)`

	broadcastIsAuthorQuery = `
		SELECT COUNT(1) 
		FROM broadcast_messages 
		WHERE id = $1 AND author_id = $2 AND deleted_at IS NULL`
)
