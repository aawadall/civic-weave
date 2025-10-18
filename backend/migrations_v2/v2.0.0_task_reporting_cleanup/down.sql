-- Rollback for Task Reporting and Legacy Column Cleanup Migration

-- Restore legacy columns
ALTER TABLE projects 
ADD COLUMN required_skills JSONB DEFAULT '[]',
ADD COLUMN status VARCHAR(50) DEFAULT 'draft';

-- Recreate dropped indexes
CREATE INDEX idx_projects_required_skills ON projects USING GIN (required_skills);
CREATE INDEX idx_projects_status ON projects (status);

-- Drop new columns from project_tasks
ALTER TABLE project_tasks 
DROP COLUMN IF EXISTS started_at,
DROP COLUMN IF EXISTS blocked_at,
DROP COLUMN IF EXISTS blocked_reason,
DROP COLUMN IF EXISTS completed_at,
DROP COLUMN IF EXISTS completion_note,
DROP COLUMN IF EXISTS takeover_requested_at,
DROP COLUMN IF EXISTS takeover_reason,
DROP COLUMN IF EXISTS last_status_changed_by;

-- Drop task_activity_log table and indexes
DROP INDEX IF EXISTS idx_task_activity_log_actor_user_id;
DROP INDEX IF EXISTS idx_task_activity_log_created_at;
DROP INDEX IF EXISTS idx_task_activity_log_task_id;
DROP TABLE IF EXISTS task_activity_log;
