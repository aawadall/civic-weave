-- UP
-- Task Management Enhancements Migration
-- Adds task comments, time logging, new statuses, and task-related messaging

-- ============================================================================
-- PART 1: New Tables for Task Comments and Time Logging
-- ============================================================================

-- Create task_comments table (replaces simple task_updates with richer comments)
CREATE TABLE task_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES project_tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    comment_text TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    edited_at TIMESTAMP
);

-- Create task_time_logs table for tracking volunteer hours
CREATE TABLE task_time_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES project_tasks(id) ON DELETE CASCADE,
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    hours DECIMAL(5,2) NOT NULL CHECK (hours > 0),
    log_date DATE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- PART 2: Modify Existing Tables
-- ============================================================================

-- Add new status values to project_tasks
ALTER TABLE project_tasks 
DROP CONSTRAINT IF EXISTS project_tasks_status_check;

ALTER TABLE project_tasks 
ADD CONSTRAINT project_tasks_status_check 
CHECK (status IN ('todo', 'in_progress', 'done', 'blocked', 'takeover_requested'));

-- Add task_id and message_type to project_messages for task-related notifications
ALTER TABLE project_messages 
ADD COLUMN IF NOT EXISTS task_id UUID REFERENCES project_tasks(id) ON DELETE CASCADE;

ALTER TABLE project_messages 
ADD COLUMN IF NOT EXISTS message_type VARCHAR(50) DEFAULT 'general' 
CHECK (message_type IN ('general', 'task_done', 'task_blocked', 'task_takeover'));

-- ============================================================================
-- PART 3: Indexes for Performance
-- ============================================================================

-- Task comments indexes
CREATE INDEX idx_task_comments_task_id ON task_comments(task_id);
CREATE INDEX idx_task_comments_user_id ON task_comments(user_id);
CREATE INDEX idx_task_comments_created_at ON task_comments(created_at);

-- Task time logs indexes
CREATE INDEX idx_task_time_logs_task_id ON task_time_logs(task_id);
CREATE INDEX idx_task_time_logs_volunteer_id ON task_time_logs(volunteer_id);
CREATE INDEX idx_task_time_logs_log_date ON task_time_logs(log_date);

-- Project messages indexes for task-related messages
CREATE INDEX idx_project_messages_task_id ON project_messages(task_id);
CREATE INDEX idx_project_messages_type ON project_messages(message_type);

-- ============================================================================
-- PART 4: Helper Functions for Time Aggregation
-- ============================================================================

-- Function to get total hours for a specific task
CREATE OR REPLACE FUNCTION get_task_total_hours(task_uuid UUID)
RETURNS DECIMAL(5,2) AS $$
BEGIN
    RETURN COALESCE(
        (SELECT SUM(hours) FROM task_time_logs WHERE task_id = task_uuid),
        0.00
    );
END;
$$ LANGUAGE plpgsql;

-- Function to get total hours for a volunteer in a project
CREATE OR REPLACE FUNCTION get_volunteer_project_hours(volunteer_uuid UUID, project_uuid UUID)
RETURNS DECIMAL(5,2) AS $$
BEGIN
    RETURN COALESCE(
        (SELECT SUM(ttl.hours) 
         FROM task_time_logs ttl
         JOIN project_tasks pt ON ttl.task_id = pt.id
         WHERE ttl.volunteer_id = volunteer_uuid 
           AND pt.project_id = project_uuid),
        0.00
    );
END;
$$ LANGUAGE plpgsql;

-- Function to get total hours for a project (all tasks)
CREATE OR REPLACE FUNCTION get_project_total_hours(project_uuid UUID)
RETURNS DECIMAL(5,2) AS $$
BEGIN
    RETURN COALESCE(
        (SELECT SUM(ttl.hours) 
         FROM task_time_logs ttl
         JOIN project_tasks pt ON ttl.task_id = pt.id
         WHERE pt.project_id = project_uuid),
        0.00
    );
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- PART 5: Update Triggers for Timestamps
-- ============================================================================

-- Create trigger for task_comments updated_at
CREATE TRIGGER update_task_comments_updated_at 
BEFORE UPDATE ON task_comments 
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- DOWN
-- ============================================================================

-- Drop triggers
DROP TRIGGER IF EXISTS update_task_comments_updated_at ON task_comments;

-- Drop functions
DROP FUNCTION IF EXISTS get_task_total_hours(UUID);
DROP FUNCTION IF EXISTS get_volunteer_project_hours(UUID, UUID);
DROP FUNCTION IF EXISTS get_project_total_hours(UUID);

-- Drop indexes
DROP INDEX IF EXISTS idx_project_messages_type;
DROP INDEX IF EXISTS idx_project_messages_task_id;
DROP INDEX IF EXISTS idx_task_time_logs_log_date;
DROP INDEX IF EXISTS idx_task_time_logs_volunteer_id;
DROP INDEX IF EXISTS idx_task_time_logs_task_id;
DROP INDEX IF EXISTS idx_task_comments_created_at;
DROP INDEX IF EXISTS idx_task_comments_user_id;
DROP INDEX IF EXISTS idx_task_comments_task_id;

-- Remove columns from project_messages
ALTER TABLE project_messages DROP COLUMN IF EXISTS message_type;
ALTER TABLE project_messages DROP COLUMN IF EXISTS task_id;

-- Restore original status constraint
ALTER TABLE project_tasks 
DROP CONSTRAINT IF EXISTS project_tasks_status_check;

ALTER TABLE project_tasks 
ADD CONSTRAINT project_tasks_status_check 
CHECK (status IN ('todo', 'in_progress', 'done'));

-- Drop tables
DROP TABLE IF EXISTS task_time_logs;
DROP TABLE IF EXISTS task_comments;
