-- UP
-- Project Module Enhancements Migration
-- Adds tasks, messaging, WYSIWYG content, and budget tracking

-- ============================================================================
-- PART 1: Project Tasks System
-- ============================================================================

-- Create project_tasks table
CREATE TABLE project_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    assignee_id UUID REFERENCES volunteers(id) ON DELETE SET NULL,
    created_by_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'todo' CHECK (status IN ('todo', 'in_progress', 'done')),
    priority VARCHAR(10) DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high')),
    due_date TIMESTAMP,
    labels JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create task_updates table for progress tracking
CREATE TABLE task_updates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES project_tasks(id) ON DELETE CASCADE,
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    update_text TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for performance
CREATE INDEX idx_project_tasks_project_id ON project_tasks(project_id);
CREATE INDEX idx_project_tasks_assignee_id ON project_tasks(assignee_id);
CREATE INDEX idx_project_tasks_status ON project_tasks(status);
CREATE INDEX idx_project_tasks_due_date ON project_tasks(due_date);
CREATE INDEX idx_project_tasks_created_by ON project_tasks(created_by_id);

CREATE INDEX idx_task_updates_task_id ON task_updates(task_id);
CREATE INDEX idx_task_updates_volunteer_id ON task_updates(volunteer_id);

-- ============================================================================
-- PART 2: Messaging System
-- ============================================================================

-- Create project_messages table
CREATE TABLE project_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message_text TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    edited_at TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create message_reads table for read tracking
CREATE TABLE message_reads (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message_id UUID NOT NULL REFERENCES project_messages(id) ON DELETE CASCADE,
    read_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, message_id)
);

-- Add indexes for performance
CREATE INDEX idx_project_messages_project_id ON project_messages(project_id);
CREATE INDEX idx_project_messages_sender_id ON project_messages(sender_id);
CREATE INDEX idx_project_messages_created_at ON project_messages(created_at);
CREATE INDEX idx_project_messages_deleted_at ON project_messages(deleted_at);

CREATE INDEX idx_message_reads_user_id ON message_reads(user_id);
CREATE INDEX idx_message_reads_message_id ON message_reads(message_id);

-- ============================================================================
-- PART 3: Enhanced Projects Schema
-- ============================================================================

-- Add new columns to projects table
ALTER TABLE projects ADD COLUMN IF NOT EXISTS content_json JSONB DEFAULT NULL;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS budget_total DECIMAL(12, 2) DEFAULT 0.00;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS budget_spent DECIMAL(12, 2) DEFAULT 0.00;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS permissions JSONB DEFAULT '{}';

-- Add comment to explain content_json format
COMMENT ON COLUMN projects.content_json IS 'TipTap editor JSON format for rich text project description';

-- ============================================================================
-- PART 4: Application Auto-Enrollment Trigger
-- ============================================================================

-- Create function to auto-enroll volunteers when application is accepted
CREATE OR REPLACE FUNCTION auto_enroll_volunteer()
RETURNS TRIGGER AS $$
BEGIN
    -- If status changed to 'accepted', add to project_team_members
    IF NEW.status = 'accepted' AND (OLD.status IS NULL OR OLD.status != 'accepted') THEN
        -- Insert into project_team_members if not already exists
        INSERT INTO project_team_members (project_id, volunteer_id, status, joined_at)
        VALUES (NEW.project_id, NEW.volunteer_id, 'active', CURRENT_TIMESTAMP)
        ON CONFLICT (project_id, volunteer_id) DO UPDATE
        SET status = 'active', updated_at = CURRENT_TIMESTAMP;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger on applications table
CREATE TRIGGER trigger_auto_enroll_volunteer
AFTER UPDATE ON applications
FOR EACH ROW
WHEN (NEW.status = 'accepted')
EXECUTE FUNCTION auto_enroll_volunteer();

-- ============================================================================
-- PART 5: Update triggers for timestamps
-- ============================================================================

CREATE TRIGGER update_project_tasks_updated_at 
BEFORE UPDATE ON project_tasks 
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- DOWN
-- ============================================================================

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_auto_enroll_volunteer ON applications;
DROP FUNCTION IF EXISTS auto_enroll_volunteer();

DROP TRIGGER IF EXISTS update_project_tasks_updated_at ON project_tasks;

-- Drop indexes
DROP INDEX IF EXISTS idx_message_reads_message_id;
DROP INDEX IF EXISTS idx_message_reads_user_id;
DROP INDEX IF EXISTS idx_project_messages_deleted_at;
DROP INDEX IF EXISTS idx_project_messages_created_at;
DROP INDEX IF EXISTS idx_project_messages_sender_id;
DROP INDEX IF EXISTS idx_project_messages_project_id;

DROP INDEX IF EXISTS idx_task_updates_volunteer_id;
DROP INDEX IF EXISTS idx_task_updates_task_id;
DROP INDEX IF EXISTS idx_project_tasks_created_by;
DROP INDEX IF EXISTS idx_project_tasks_due_date;
DROP INDEX IF EXISTS idx_project_tasks_status;
DROP INDEX IF EXISTS idx_project_tasks_assignee_id;
DROP INDEX IF EXISTS idx_project_tasks_project_id;

-- Drop tables
DROP TABLE IF EXISTS message_reads;
DROP TABLE IF EXISTS project_messages;
DROP TABLE IF EXISTS task_updates;
DROP TABLE IF EXISTS project_tasks;

-- Remove columns from projects
ALTER TABLE projects DROP COLUMN IF EXISTS permissions;
ALTER TABLE projects DROP COLUMN IF EXISTS budget_spent;
ALTER TABLE projects DROP COLUMN IF EXISTS budget_total;
ALTER TABLE projects DROP COLUMN IF EXISTS content_json;

