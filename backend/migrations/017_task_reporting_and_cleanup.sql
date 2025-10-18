-- UP
-- Task Reporting and Legacy Column Cleanup Migration
-- Adds timeline tracking to tasks, creates activity log, and removes legacy project columns

-- ============================================================================
-- PART 1: Add Timeline Columns to project_tasks
-- ============================================================================

-- Add timeline tracking columns to project_tasks
ALTER TABLE project_tasks 
ADD COLUMN started_at TIMESTAMP,
ADD COLUMN blocked_at TIMESTAMP,
ADD COLUMN blocked_reason TEXT,
ADD COLUMN completed_at TIMESTAMP,
ADD COLUMN completion_note TEXT,
ADD COLUMN takeover_requested_at TIMESTAMP,
ADD COLUMN takeover_reason TEXT,
ADD COLUMN last_status_changed_by UUID REFERENCES users(id);

-- ============================================================================
-- PART 2: Create task_activity_log table
-- ============================================================================

-- Create task_activity_log table for tracking task status changes
CREATE TABLE task_activity_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES project_tasks(id) ON DELETE CASCADE,
    actor_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_volunteer_id UUID REFERENCES volunteers(id) ON DELETE CASCADE,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    context JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_task_activity_log_task_id ON task_activity_log(task_id);
CREATE INDEX idx_task_activity_log_created_at ON task_activity_log(created_at);
CREATE INDEX idx_task_activity_log_actor_user_id ON task_activity_log(actor_user_id);

-- ============================================================================
-- PART 3: Back-fill task_activity_log with existing data
-- ============================================================================

-- Insert initial activity log entries for existing tasks
-- This creates a baseline entry for each task using the task creation time
INSERT INTO task_activity_log (task_id, actor_user_id, from_status, to_status, context, created_at)
SELECT 
    pt.id as task_id,
    pt.created_by_id as actor_user_id,
    NULL as from_status,
    pt.status as to_status,
    '{"initial_creation": true}'::jsonb as context,
    pt.created_at
FROM project_tasks pt
WHERE pt.created_at IS NOT NULL;

-- ============================================================================
-- PART 4: Drop legacy columns from projects table
-- ============================================================================

-- Drop legacy columns and indexes
DROP INDEX IF EXISTS idx_projects_required_skills;
DROP INDEX IF EXISTS idx_projects_status;

-- Drop the legacy columns
ALTER TABLE projects DROP COLUMN IF EXISTS required_skills;
ALTER TABLE projects DROP COLUMN IF EXISTS status;

-- ============================================================================
-- PART 5: Update triggers for project_tasks.updated_at
-- ============================================================================

-- Ensure the updated_at trigger works with new columns
-- The existing trigger should already handle this, but we'll verify it exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger 
        WHERE tgname = 'update_project_tasks_updated_at'
    ) THEN
        CREATE TRIGGER update_project_tasks_updated_at 
        BEFORE UPDATE ON project_tasks 
        FOR EACH ROW 
        EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

-- ============================================================================
-- DOWN
-- ============================================================================

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
