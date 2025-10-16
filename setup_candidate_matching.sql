-- =============================================================================
-- CivicWeave Candidate Matching System - Complete Setup Script
-- =============================================================================
-- This script creates the candidate_notifications tracking table and 
-- includes verification queries to test the system.
--
-- Usage: psql -h localhost -U postgres -d civicweave -f setup_candidate_matching.sql
-- =============================================================================

\echo '=================================================='
\echo 'CivicWeave Candidate Matching System Setup'
\echo '=================================================='
\echo ''

-- =============================================================================
-- 1. CREATE CANDIDATE_NOTIFICATIONS TABLE
-- =============================================================================
\echo '1. Creating candidate_notifications table...'

CREATE TABLE IF NOT EXISTS candidate_notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    match_score DECIMAL(5,4) NOT NULL,
    notified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    notification_batch_id UUID NOT NULL,
    UNIQUE(project_id, volunteer_id, notification_batch_id)
);

\echo '   ✓ Table created'

-- =============================================================================
-- 2. CREATE INDEXES FOR PERFORMANCE
-- =============================================================================
\echo '2. Creating indexes...'

CREATE INDEX IF NOT EXISTS idx_candidate_notifications_project_id 
    ON candidate_notifications(project_id);
    
CREATE INDEX IF NOT EXISTS idx_candidate_notifications_volunteer_id 
    ON candidate_notifications(volunteer_id);
    
CREATE INDEX IF NOT EXISTS idx_candidate_notifications_batch_id 
    ON candidate_notifications(notification_batch_id);
    
CREATE INDEX IF NOT EXISTS idx_candidate_notifications_notified_at 
    ON candidate_notifications(notified_at);

\echo '   ✓ Indexes created'
\echo ''

-- =============================================================================
-- 3. VERIFICATION
-- =============================================================================
\echo '=================================================='
\echo 'Verification'
\echo '=================================================='
\echo ''

\echo '3. Table structure:'
\d candidate_notifications

\echo ''
\echo '4. Indexes:'
\di candidate_notifications*

\echo ''
\echo '=================================================='
\echo 'System Status'
\echo '=================================================='
\echo ''

-- Check recruiting projects
\echo '5. Recruiting/Active Projects:'
SELECT 
    COUNT(*) as recruiting_projects,
    SUM(CASE WHEN project_status = 'recruiting' THEN 1 ELSE 0 END) as recruiting,
    SUM(CASE WHEN project_status = 'active' THEN 1 ELSE 0 END) as active
FROM projects
WHERE project_status IN ('recruiting', 'active');

\echo ''
\echo '6. Projects with Team Leads:'
SELECT 
    COUNT(*) as total_projects,
    COUNT(team_lead_id) as projects_with_team_lead,
    COUNT(*) - COUNT(team_lead_id) as projects_without_team_lead
FROM projects
WHERE project_status IN ('recruiting', 'active');

\echo ''
\echo '7. Volunteer Match Coverage:'
SELECT 
    COUNT(DISTINCT initiative_id) as projects_with_matches,
    COUNT(DISTINCT volunteer_id) as volunteers_with_matches,
    COUNT(*) as total_match_records,
    ROUND(AVG(match_score)::numeric, 3) as avg_match_score
FROM volunteer_initiative_matches
WHERE match_score >= 0.6;

\echo ''
\echo '8. Visible Volunteers:'
SELECT 
    COUNT(*) as total_volunteers,
    SUM(CASE WHEN skills_visible = true THEN 1 ELSE 0 END) as visible_volunteers,
    SUM(CASE WHEN skills_visible = false THEN 1 ELSE 0 END) as hidden_volunteers
FROM volunteers;

\echo ''
\echo '=================================================='
\echo 'Sample Data Preview'
\echo '=================================================='
\echo ''

-- Show sample of eligible projects
\echo '9. Sample Recruiting Projects (Top 5):'
SELECT 
    id,
    title,
    project_status,
    CASE 
        WHEN team_lead_id IS NOT NULL THEN 'Has Team Lead'
        ELSE 'No Team Lead'
    END as lead_status,
    created_at::date
FROM projects
WHERE project_status IN ('recruiting', 'active')
ORDER BY created_at DESC
LIMIT 5;

\echo ''
\echo '10. Sample Candidates Per Project (First Project):'
WITH first_project AS (
    SELECT id 
    FROM projects 
    WHERE project_status IN ('recruiting', 'active')
    LIMIT 1
)
SELECT 
    v.name,
    v.skills_visible,
    ROUND((m.match_score * 100)::numeric, 0) || '%' as match_percentage,
    m.matched_skill_count as matched_skills
FROM volunteer_initiative_matches m
JOIN volunteers v ON m.volunteer_id = v.id
JOIN first_project p ON m.initiative_id = p.id
WHERE m.match_score >= 0.6
    AND v.skills_visible = true
ORDER BY m.match_score DESC
LIMIT 5;

\echo ''
\echo '=================================================='
\echo 'Useful Monitoring Queries'
\echo '=================================================='
\echo ''
\echo 'Run these queries after the batch job executes:'
\echo ''
\echo '-- Check recent notifications'
\echo 'SELECT COUNT(*) FROM candidate_notifications WHERE notified_at > NOW() - INTERVAL ''1 hour'';'
\echo ''
\echo '-- View notification activity by date'
\echo 'SELECT DATE(notified_at) as date, COUNT(*) as notifications'
\echo 'FROM candidate_notifications'
\echo 'GROUP BY DATE(notified_at)'
\echo 'ORDER BY date DESC;'
\echo ''
\echo '-- Check project messages (notifications)'
\echo 'SELECT * FROM project_messages'
\echo 'WHERE created_at > NOW() - INTERVAL ''1 hour'''
\echo 'ORDER BY created_at DESC;'
\echo ''

\echo '=================================================='
\echo 'Setup Complete!'
\echo '=================================================='
\echo ''
\echo 'Next Steps:'
\echo '1. Configure environment variables in backend/.env:'
\echo '   TOP_K_CANDIDATES=10'
\echo '   MIN_MATCH_SCORE=0.6'
\echo '   SYSTEM_USER_ID=00000000-0000-0000-0000-000000000000'
\echo ''
\echo '2. Set up Python environment:'
\echo '   cd backend && python3 -m venv venv'
\echo '   source venv/bin/activate'
\echo '   pip install psycopg2-binary python-dotenv'
\echo ''
\echo '3. Run match calculation:'
\echo '   make job-calculate-matches'
\echo ''
\echo '4. Run notification job:'
\echo '   make job-notify-matches'
\echo ''
\echo '5. Schedule daily execution:'
\echo '   crontab -e'
\echo '   # Add: 0 9 * * * cd /path/to/CivicWeave/backend && ./jobs/run_daily_matching.sh'
\echo ''
\echo '=================================================='



