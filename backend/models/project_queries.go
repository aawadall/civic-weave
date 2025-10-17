package models

// Query constants for ProjectService
const (
	projectCreateQuery = `
		INSERT INTO projects (id, title, description, required_skills, location_lat, location_lng, 
		                     location_address, start_date, end_date, project_status, 
		                     created_by_admin_id, team_lead_id, auto_notify_matches)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at, updated_at`

	projectGetByIDQuery = `
		SELECT id, title, description, required_skills, location_lat, location_lng, 
		       location_address, start_date, end_date, project_status, 
		       created_by_admin_id, team_lead_id, auto_notify_matches, created_at, updated_at
		FROM projects WHERE id = $1`

	projectListWithSkillsQuery = `
		SELECT p.id, p.title, p.description, p.location_lat, p.location_lng, 
		       p.location_address, p.start_date, p.end_date, p.project_status, 
		       p.created_by_admin_id, p.team_lead_id, p.auto_notify_matches, p.created_at, p.updated_at,
		       CASE 
		           WHEN COUNT(prs.skill_id) > 0 THEN
		               JSON_AGG(
		                   JSON_BUILD_OBJECT('id', st.id, 'name', st.skill_name)
		               ) FILTER (WHERE st.id IS NOT NULL)::jsonb
		           ELSE COALESCE(p.required_skills, '[]'::jsonb)
		       END as required_skills
		FROM projects p
		LEFT JOIN project_required_skills prs ON p.id = prs.project_id
		LEFT JOIN skill_taxonomy st ON prs.skill_id = st.id
		WHERE ($1::text IS NULL OR p.project_status::text = $1)
		GROUP BY p.id, p.title, p.description, p.location_lat, p.location_lng, 
		         p.location_address, p.start_date, p.end_date, p.project_status, 
		         p.created_by_admin_id, p.team_lead_id, p.auto_notify_matches, p.created_at, p.updated_at,
		         p.required_skills
		HAVING ($2::jsonb IS NULL OR $2::jsonb = '[]'::jsonb OR 
		        EXISTS (
		            SELECT 1 FROM project_required_skills prs2 
		            JOIN skill_taxonomy st2 ON prs2.skill_id = st2.id 
		            WHERE prs2.project_id = p.id AND st2.skill_name = ANY($2::text[])
		        ) OR
		        EXISTS (
		            SELECT 1 FROM jsonb_array_elements_text(p.required_skills) AS skill
		            WHERE skill = ANY($2::text[])
		        ))
		ORDER BY p.created_at DESC
		LIMIT $3 OFFSET $4`

	projectListQuery = `
		SELECT p.id, p.title, p.description, p.location_lat, p.location_lng, 
		       p.location_address, p.start_date, p.end_date, p.project_status, 
		       p.created_by_admin_id, p.team_lead_id, p.auto_notify_matches, p.created_at, p.updated_at,
		       CASE 
		           WHEN COUNT(prs.skill_id) > 0 THEN
		               JSON_AGG(
		                   JSON_BUILD_OBJECT('id', st.id, 'name', st.skill_name)
		               ) FILTER (WHERE st.id IS NOT NULL)::jsonb
		           ELSE COALESCE(p.required_skills, '[]'::jsonb)
		       END as required_skills
		FROM projects p
		LEFT JOIN project_required_skills prs ON p.id = prs.project_id
		LEFT JOIN skill_taxonomy st ON prs.skill_id = st.id
		WHERE ($1::text IS NULL OR p.project_status::text = $1)
		GROUP BY p.id, p.title, p.description, p.location_lat, p.location_lng, 
		         p.location_address, p.start_date, p.end_date, p.project_status, 
		         p.created_by_admin_id, p.team_lead_id, p.auto_notify_matches, p.created_at, p.updated_at,
		         p.required_skills
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3`

	projectListByTeamLeadQuery = `
		SELECT p.id, p.title, p.description, p.location_lat, p.location_lng, 
		       p.location_address, p.start_date, p.end_date, p.project_status, 
		       p.created_by_admin_id, p.team_lead_id, p.auto_notify_matches, p.created_at, p.updated_at,
		       CASE 
		           WHEN COUNT(prs.skill_id) > 0 THEN
		               JSON_AGG(
		                   JSON_BUILD_OBJECT('id', st.id, 'name', st.skill_name)
		               ) FILTER (WHERE st.id IS NOT NULL)::jsonb
		           ELSE COALESCE(p.required_skills, '[]'::jsonb)
		       END as required_skills
		FROM projects p
		LEFT JOIN project_required_skills prs ON p.id = prs.project_id
		LEFT JOIN skill_taxonomy st ON prs.skill_id = st.id
		WHERE p.team_lead_id = $1
		GROUP BY p.id, p.title, p.description, p.location_lat, p.location_lng, 
		         p.location_address, p.start_date, p.end_date, p.project_status, 
		         p.created_by_admin_id, p.team_lead_id, p.auto_notify_matches, p.created_at, p.updated_at,
		         p.required_skills
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3`

	projectUpdateQuery = `
		UPDATE projects 
		SET title = $2, description = $3, required_skills = $4, location_lat = $5, 
		    location_lng = $6, location_address = $7, start_date = $8, end_date = $9, 
		    project_status = $10, team_lead_id = $11, auto_notify_matches = $12, 
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
		SELECT p.id, p.title, p.description, p.location_lat, p.location_lng, 
		       p.location_address, p.start_date, p.end_date, p.project_status, 
		       p.created_by_admin_id, p.team_lead_id, p.auto_notify_matches, p.created_at, p.updated_at,
		       CASE 
		           WHEN COUNT(prs.skill_id) > 0 THEN
		               JSON_AGG(
		                   JSON_BUILD_OBJECT('id', st.id, 'name', st.skill_name)
		               ) FILTER (WHERE st.id IS NOT NULL)::jsonb
		           ELSE COALESCE(p.required_skills, '[]'::jsonb)
		       END as required_skills
		FROM projects p
		LEFT JOIN project_required_skills prs ON p.id = prs.project_id
		LEFT JOIN skill_taxonomy st ON prs.skill_id = st.id
		WHERE p.project_status IN ('recruiting', 'active')
		GROUP BY p.id, p.title, p.description, p.location_lat, p.location_lng, 
		         p.location_address, p.start_date, p.end_date, p.project_status, 
		         p.created_by_admin_id, p.team_lead_id, p.auto_notify_matches, p.created_at, p.updated_at,
		         p.required_skills
		ORDER BY p.created_at DESC`

	projectIsCreatorQuery = `SELECT COUNT(1) FROM projects WHERE id = $1 AND created_by_admin_id = $2`

	projectTransitionStatusQuery = `
		UPDATE projects 
		SET project_status = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	projectActiveTeamCountQuery = `
		SELECT COUNT(1) 
		FROM project_team_members 
		WHERE project_id = $1 AND status = 'active'`
)
