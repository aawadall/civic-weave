package models

// Query constants for ProjectService
const (
	projectCreateQuery = `
		INSERT INTO projects (id, title, description, required_skills, location_lat, location_lng, 
		                     location_address, start_date, end_date, status, project_status, 
		                     created_by_admin_id, team_lead_id, auto_notify_matches)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING created_at, updated_at`

	projectGetByIDQuery = `
		SELECT id, title, description, required_skills, location_lat, location_lng, 
		       location_address, start_date, end_date, status, project_status, 
		       created_by_admin_id, team_lead_id, auto_notify_matches, created_at, updated_at
		FROM projects WHERE id = $1`

	projectListWithSkillsQuery = `
		SELECT id, title, description, required_skills, location_lat, location_lng, 
		       location_address, start_date, end_date, status, project_status, 
		       created_by_admin_id, team_lead_id, auto_notify_matches, created_at, updated_at
		FROM projects
		WHERE ($1::text IS NULL OR status = $1 OR project_status::text = $1)
		AND required_skills && $2::jsonb
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	projectListQuery = `
		SELECT id, title, description, required_skills, location_lat, location_lng, 
		       location_address, start_date, end_date, status, project_status, 
		       created_by_admin_id, team_lead_id, auto_notify_matches, created_at, updated_at
		FROM projects
		WHERE ($1::text IS NULL OR status = $1 OR project_status::text = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	projectListByTeamLeadQuery = `
		SELECT id, title, description, required_skills, location_lat, location_lng, 
		       location_address, start_date, end_date, status, project_status, 
		       created_by_admin_id, team_lead_id, auto_notify_matches, created_at, updated_at
		FROM projects
		WHERE team_lead_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	projectUpdateQuery = `
		UPDATE projects 
		SET title = $2, description = $3, required_skills = $4, location_lat = $5, 
		    location_lng = $6, location_address = $7, start_date = $8, end_date = $9, 
		    status = $10, project_status = $11, team_lead_id = $12, auto_notify_matches = $13, 
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	projectDeleteQuery = `DELETE FROM projects WHERE id = $1`

	projectGetTeamMembersQuery = `
		SELECT id, project_id, volunteer_id, joined_at, status, created_at, updated_at
		FROM project_team_members 
		WHERE project_id = $1
		ORDER BY joined_at`

	projectAddTeamMemberQuery = `
		INSERT INTO project_team_members (id, project_id, volunteer_id, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (project_id, volunteer_id) DO UPDATE SET 
		status = EXCLUDED.status, updated_at = CURRENT_TIMESTAMP`

	projectUpdateTeamMemberStatusQuery = `
		UPDATE project_team_members 
		SET status = $3, updated_at = CURRENT_TIMESTAMP
		WHERE project_id = $1 AND volunteer_id = $2`

	projectRemoveTeamMemberQuery = `DELETE FROM project_team_members WHERE project_id = $1 AND volunteer_id = $2`

	projectIsTeamLeadQuery = `SELECT COUNT(1) FROM projects WHERE id = $1 AND team_lead_id = $2`

	projectIsTeamMemberQuery = `
		SELECT COUNT(1) 
		FROM project_team_members ptm
		JOIN volunteers v ON ptm.volunteer_id = v.id
		WHERE ptm.project_id = $1 AND v.user_id = $2 AND ptm.status = 'active'`

	projectAssignTeamLeadQuery = `
		UPDATE projects 
		SET team_lead_id = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	projectGetActiveProjectsQuery = `
		SELECT id, title, description, required_skills, location_lat, location_lng, 
		       location_address, start_date, end_date, status, project_status, 
		       created_by_admin_id, team_lead_id, auto_notify_matches, created_at, updated_at
		FROM projects
		WHERE project_status IN ('recruiting', 'active')
		ORDER BY created_at DESC`
)
