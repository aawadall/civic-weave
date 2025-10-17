package models

// Query constants for ResourceService
const (
	resourceCreateQuery = `
		INSERT INTO resources (id, title, description, resource_type, file_url, file_size, 
		                       mime_type, scope, project_id, uploaded_by_id, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at`

	resourceGetByIDQuery = `
		SELECT id, title, description, resource_type, file_url, file_size, mime_type, 
		       scope, project_id, uploaded_by_id, tags, download_count, 
		       created_at, updated_at, deleted_at
		FROM resources WHERE id = $1 AND deleted_at IS NULL`

	resourceListQuery = `
		SELECT 
			r.id, r.title, r.description, r.resource_type, r.file_url, r.file_size, 
			r.mime_type, r.scope, r.project_id, r.uploaded_by_id, r.tags, r.download_count,
			r.created_at, r.updated_at, r.deleted_at,
			COALESCE(v.name, a.name, u.email) as uploader_name,
			u.email as uploader_email,
			p.title as project_title
		FROM resources r
		JOIN users u ON r.uploaded_by_id = u.id
		LEFT JOIN volunteers v ON u.id = v.user_id
		LEFT JOIN admins a ON u.id = a.user_id
		LEFT JOIN projects p ON r.project_id = p.id`

	resourceUpdateQuery = `
		UPDATE resources 
		SET title = $2, description = $3, resource_type = $4, file_url = $5, 
		    file_size = $6, mime_type = $7, scope = $8, project_id = $9, 
		    tags = $10, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at`

	resourceSoftDeleteQuery = `
		UPDATE resources 
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL`

	resourceIncrementDownloadQuery = `
		UPDATE resources 
		SET download_count = download_count + 1
		WHERE id = $1 AND deleted_at IS NULL`

	resourceGetStatsQuery = `
		SELECT 
			COUNT(*) as total_resources,
			COUNT(CASE WHEN scope = 'global' THEN 1 END) as global_resources,
			COUNT(CASE WHEN scope = 'project_specific' THEN 1 END) as project_resources,
			COALESCE(SUM(download_count), 0) as total_downloads,
			COUNT(CASE WHEN created_at > NOW() - INTERVAL '7 days' THEN 1 END) as recent_uploads
		FROM resources 
		WHERE deleted_at IS NULL`

	resourceGetRecentQuery = `
		SELECT 
			r.id, r.title, r.description, r.resource_type, r.file_url, r.file_size, 
			r.mime_type, r.scope, r.project_id, r.uploaded_by_id, r.tags, r.download_count,
			r.created_at, r.updated_at, r.deleted_at,
			COALESCE(v.name, a.name, u.email) as uploader_name,
			u.email as uploader_email,
			p.title as project_title
		FROM resources r
		JOIN users u ON r.uploaded_by_id = u.id
		LEFT JOIN volunteers v ON u.id = v.user_id
		LEFT JOIN admins a ON u.id = a.user_id
		LEFT JOIN projects p ON r.project_id = p.id
		WHERE r.deleted_at IS NULL
		ORDER BY r.created_at DESC
		LIMIT $1`

	resourceIsUploaderQuery = `
		SELECT COUNT(1) 
		FROM resources 
		WHERE id = $1 AND uploaded_by_id = $2 AND deleted_at IS NULL`
)
