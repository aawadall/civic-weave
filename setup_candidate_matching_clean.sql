-- =============================================================================
-- CivicWeave Candidate Matching System - Setup Script
-- =============================================================================
-- Run this directly in SQL Studio, pgAdmin, DBeaver, or any SQL client
-- Database: civicweave
-- =============================================================================

-- =============================================================================
-- CREATE CANDIDATE_NOTIFICATIONS TABLE
-- =============================================================================

CREATE TABLE IF NOT EXISTS candidate_notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    match_score DECIMAL(5,4) NOT NULL,
    notified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    notification_batch_id UUID NOT NULL,
    UNIQUE(project_id, volunteer_id, notification_batch_id)
);

COMMENT ON TABLE candidate_notifications IS 'Tracks which volunteers have been notified about which projects to prevent duplicate notifications';
COMMENT ON COLUMN candidate_notifications.match_score IS 'Match score between volunteer and project (0.0-1.0)';
COMMENT ON COLUMN candidate_notifications.notification_batch_id IS 'Unique ID for each batch run to group notifications';

-- =============================================================================
-- CREATE INDEXES
-- =============================================================================

CREATE INDEX IF NOT EXISTS idx_candidate_notifications_project_id 
    ON candidate_notifications(project_id);
    
CREATE INDEX IF NOT EXISTS idx_candidate_notifications_volunteer_id 
    ON candidate_notifications(volunteer_id);
    
CREATE INDEX IF NOT EXISTS idx_candidate_notifications_batch_id 
    ON candidate_notifications(notification_batch_id);
    
CREATE INDEX IF NOT EXISTS idx_candidate_notifications_notified_at 
    ON candidate_notifications(notified_at);

-- =============================================================================
-- VERIFICATION QUERIES
-- =============================================================================
-- Run these to verify the setup was successful
-- =============================================================================

-- 1. Check table structure
SELECT 
    column_name, 
    data_type, 
    is_nullable,
    column_default
FROM information_schema.columns 
WHERE table_name = 'candidate_notifications'
ORDER BY ordinal_position;

-- 2. Check indexes
SELECT 
    indexname, 
    indexdef 
FROM pg_indexes 
WHERE tablename = 'candidate_notifications'
ORDER BY indexname;

-- 3. System Status: Recruiting Projects
SELECT 
    COUNT(*) as recruiting_projects,
    SUM(CASE WHEN project_status = 'recruiting' THEN 1 ELSE 0 END) as recruiting,
    SUM(CASE WHEN project_status = 'active' THEN 1 ELSE 0 END) as active
FROM projects
WHERE project_status IN ('recruiting', 'active');

-- 4. Projects with Team Leads
SELECT 
    COUNT(*) as total_projects,
    COUNT(team_lead_id) as projects_with_team_lead,
    COUNT(*) - COUNT(team_lead_id) as projects_without_team_lead
FROM projects
WHERE project_status IN ('recruiting', 'active');

-- 5. Volunteer Match Coverage
SELECT 
    COUNT(DISTINCT project_id) as projects_with_matches,
    COUNT(DISTINCT volunteer_id) as volunteers_with_matches,
    COUNT(*) as total_match_records,
    ROUND(AVG(match_score)::numeric, 3) as avg_match_score
FROM volunteer_project_matches
WHERE match_score >= 0.6;

-- 6. Visible Volunteers
SELECT 
    COUNT(*) as total_volunteers,
    SUM(CASE WHEN skills_visible = true THEN 1 ELSE 0 END) as visible_volunteers,
    SUM(CASE WHEN skills_visible = false THEN 1 ELSE 0 END) as hidden_volunteers
FROM volunteers;

-- 7. Sample Recruiting Projects (Top 5)
SELECT 
    id,
    title,
    project_status,
    CASE 
        WHEN team_lead_id IS NOT NULL THEN 'Has Team Lead'
        ELSE 'No Team Lead'
    END as lead_status,
    created_at::date as created_date
FROM projects
WHERE project_status IN ('recruiting', 'active')
ORDER BY created_at DESC
LIMIT 5;

-- 8. Sample Top Candidates (for first recruiting project)
WITH first_project AS (
    SELECT id, title
    FROM projects 
    WHERE project_status IN ('recruiting', 'active')
    ORDER BY created_at DESC
    LIMIT 1
)
SELECT 
    fp.title as project_title,
    v.name as volunteer_name,
    v.skills_visible,
    ROUND((m.match_score * 100)::numeric, 0) || '%' as match_percentage,
    m.matched_skill_count as matched_skills
FROM volunteer_project_matches m
JOIN volunteers v ON m.volunteer_id = v.id
CROSS JOIN first_project fp
WHERE m.project_id = fp.id
    AND m.match_score >= 0.6
    AND v.skills_visible = true
ORDER BY m.match_score DESC
LIMIT 10;

-- =============================================================================
-- MONITORING QUERIES (Run after batch job executes)
-- =============================================================================

-- Check recent notifications (last 24 hours)
SELECT 
    COUNT(*) as notifications_last_24h
FROM candidate_notifications 
WHERE notified_at > NOW() - INTERVAL '24 hours';

-- Notification activity by date (last 7 days)
SELECT 
    DATE(notified_at) as notification_date,
    COUNT(*) as notifications_sent,
    COUNT(DISTINCT project_id) as projects,
    COUNT(DISTINCT volunteer_id) as volunteers
FROM candidate_notifications
WHERE notified_at > NOW() - INTERVAL '7 days'
GROUP BY DATE(notified_at)
ORDER BY notification_date DESC;

-- Top projects by candidate notifications (last 7 days)
SELECT 
    p.title,
    p.project_status,
    COUNT(cn.id) as candidate_count,
    ROUND(AVG(cn.match_score)::numeric, 3) as avg_match_score,
    MAX(cn.notified_at) as last_notified
FROM candidate_notifications cn
JOIN projects p ON cn.project_id = p.id
WHERE cn.notified_at > NOW() - INTERVAL '7 days'
GROUP BY p.id, p.title, p.project_status
ORDER BY candidate_count DESC
LIMIT 10;

-- Top matched volunteers (last 30 days)
SELECT 
    v.name,
    COUNT(cn.id) as projects_matched,
    ROUND(AVG(cn.match_score)::numeric, 3) as avg_match_score,
    MAX(cn.notified_at) as last_notification
FROM candidate_notifications cn
JOIN volunteers v ON cn.volunteer_id = v.id
WHERE cn.notified_at > NOW() - INTERVAL '30 days'
GROUP BY v.id, v.name
ORDER BY projects_matched DESC
LIMIT 20;

-- Recent project messages (notifications appear here)
SELECT 
    pm.id,
    p.title as project_title,
    LEFT(pm.message_text, 100) as message_preview,
    pm.created_at
FROM project_messages pm
JOIN projects p ON pm.project_id = p.id
WHERE pm.created_at > NOW() - INTERVAL '24 hours'
ORDER BY pm.created_at DESC
LIMIT 20;

-- =============================================================================
-- SUCCESS!
-- =============================================================================
-- Table 'candidate_notifications' created successfully
-- All indexes created
-- 
-- Next Steps:
-- 1. Configure backend/.env with:
--    TOP_K_CANDIDATES=10
--    MIN_MATCH_SCORE=0.6
--    SYSTEM_USER_ID=00000000-0000-0000-0000-000000000000
--
-- 2. Set up Python: make job-setup-python
-- 3. Calculate matches: make job-calculate-matches
-- 4. Run notifications: make job-notify-matches
-- 5. Schedule daily: Edit backend/jobs/crontab.example and install
-- =============================================================================

