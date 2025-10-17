-- UP
-- Messaging Expansion, Broadcast System, and Resource Library Migration
-- Expands messaging beyond projects, adds broadcast announcements, and resource library

-- ============================================================================
-- PART 1: Messaging System Expansion
-- ============================================================================

-- Add new columns to project_messages table for universal messaging
ALTER TABLE project_messages ADD COLUMN IF NOT EXISTS recipient_user_id UUID REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE project_messages ADD COLUMN IF NOT EXISTS recipient_team_id UUID REFERENCES projects(id) ON DELETE CASCADE;
ALTER TABLE project_messages ADD COLUMN IF NOT EXISTS message_scope VARCHAR(20) DEFAULT 'project' CHECK (message_scope IN ('user_to_user', 'user_to_team', 'broadcast', 'project'));
ALTER TABLE project_messages ADD COLUMN IF NOT EXISTS subject VARCHAR(255);
ALTER TABLE project_messages ADD COLUMN IF NOT EXISTS task_id UUID REFERENCES project_tasks(id) ON DELETE CASCADE;

-- Add constraint: at least one of (project_id, recipient_user_id, recipient_team_id) must be set
ALTER TABLE project_messages ADD CONSTRAINT check_message_recipient 
CHECK (
  (project_id IS NOT NULL) OR 
  (recipient_user_id IS NOT NULL) OR 
  (recipient_team_id IS NOT NULL)
);

-- Add indexes for new messaging features
CREATE INDEX IF NOT EXISTS idx_project_messages_recipient_user ON project_messages(recipient_user_id);
CREATE INDEX IF NOT EXISTS idx_project_messages_recipient_team ON project_messages(recipient_team_id);
CREATE INDEX IF NOT EXISTS idx_project_messages_scope ON project_messages(message_scope);
CREATE INDEX IF NOT EXISTS idx_project_messages_subject ON project_messages(subject);

-- ============================================================================
-- PART 2: Broadcast Messages System
-- ============================================================================

-- Create broadcast_messages table
CREATE TABLE broadcast_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_audience VARCHAR(20) NOT NULL CHECK (target_audience IN ('all_users', 'volunteers_only', 'admins_only', 'team_leads_only')),
    priority VARCHAR(10) NOT NULL DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create broadcast_reads table for read tracking
CREATE TABLE broadcast_reads (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    broadcast_id UUID NOT NULL REFERENCES broadcast_messages(id) ON DELETE CASCADE,
    read_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, broadcast_id)
);

-- Add indexes for broadcast system
CREATE INDEX idx_broadcast_messages_author ON broadcast_messages(author_id);
CREATE INDEX idx_broadcast_messages_target_audience ON broadcast_messages(target_audience);
CREATE INDEX idx_broadcast_messages_priority ON broadcast_messages(priority);
CREATE INDEX idx_broadcast_messages_expires_at ON broadcast_messages(expires_at);
CREATE INDEX idx_broadcast_messages_created_at ON broadcast_messages(created_at);
CREATE INDEX idx_broadcast_messages_deleted_at ON broadcast_messages(deleted_at);

CREATE INDEX idx_broadcast_reads_user_id ON broadcast_reads(user_id);
CREATE INDEX idx_broadcast_reads_broadcast_id ON broadcast_reads(broadcast_id);

-- ============================================================================
-- PART 3: Resource Library System
-- ============================================================================

-- Create resources table
CREATE TABLE resources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    resource_type VARCHAR(20) NOT NULL CHECK (resource_type IN ('file', 'link', 'document')),
    file_url TEXT NOT NULL,
    file_size BIGINT,
    mime_type VARCHAR(100),
    scope VARCHAR(20) NOT NULL DEFAULT 'global' CHECK (scope IN ('global', 'project_specific')),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    uploaded_by_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tags JSONB DEFAULT '[]',
    download_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Add constraint: project_id required for project_specific scope
ALTER TABLE resources ADD CONSTRAINT check_resource_scope 
CHECK (
  (scope = 'global') OR 
  (scope = 'project_specific' AND project_id IS NOT NULL)
);

-- Add indexes for resource library
CREATE INDEX idx_resources_scope ON resources(scope);
CREATE INDEX idx_resources_project_id ON resources(project_id);
CREATE INDEX idx_resources_uploaded_by ON resources(uploaded_by_id);
CREATE INDEX idx_resources_resource_type ON resources(resource_type);
CREATE INDEX idx_resources_tags ON resources USING GIN(tags);
CREATE INDEX idx_resources_created_at ON resources(created_at);
CREATE INDEX idx_resources_download_count ON resources(download_count);
CREATE INDEX idx_resources_deleted_at ON resources(deleted_at);

-- ============================================================================
-- PART 4: User Dashboard Data Views
-- ============================================================================

-- Create view for user enrolled projects with stats
CREATE OR REPLACE VIEW user_enrolled_projects AS
SELECT 
    p.id,
    p.title,
    p.description,
    p.project_status,
    p.start_date,
    p.end_date,
    p.location_address,
    ptm.joined_at,
    ptm.status as membership_status,
    v.user_id,
    CASE 
        WHEN p.team_lead_id = v.id THEN 'team_lead'
        ELSE 'member'
    END as user_role,
    COALESCE(msg_stats.unread_count, 0) as unread_message_count,
    COALESCE(task_stats.assigned_tasks, 0) as assigned_tasks_count,
    COALESCE(task_stats.overdue_tasks, 0) as overdue_tasks_count
FROM projects p
JOIN project_team_members ptm ON p.id = ptm.project_id
JOIN volunteers v ON ptm.volunteer_id = v.id
LEFT JOIN (
    SELECT 
        pm.project_id,
        COUNT(*) as unread_count
    FROM project_messages pm
    LEFT JOIN message_reads mr ON pm.id = mr.message_id
    WHERE pm.deleted_at IS NULL 
    AND mr.user_id IS NULL
    GROUP BY pm.project_id
) msg_stats ON p.id = msg_stats.project_id
LEFT JOIN (
    SELECT 
        pt.project_id,
        COUNT(*) as assigned_tasks,
        COUNT(CASE WHEN pt.due_date < NOW() AND pt.status != 'done' THEN 1 END) as overdue_tasks
    FROM project_tasks pt
    GROUP BY pt.project_id
) task_stats ON p.id = task_stats.project_id
WHERE ptm.status = 'active';

-- Create view for user assigned tasks across all projects
CREATE OR REPLACE VIEW user_assigned_tasks AS
SELECT 
    pt.id,
    pt.title,
    pt.description,
    pt.status,
    pt.priority,
    pt.due_date,
    pt.labels,
    pt.created_at,
    pt.updated_at,
    p.id as project_id,
    p.title as project_title,
    p.project_status,
    v.name as assignee_name,
    u.email as assignee_email
FROM project_tasks pt
JOIN projects p ON pt.project_id = p.id
JOIN volunteers v ON pt.assignee_id = v.id
JOIN users u ON v.user_id = u.id
;

-- ============================================================================
-- PART 5: Update triggers for timestamps
-- ============================================================================

-- Create trigger for broadcast_messages updated_at
CREATE TRIGGER update_broadcast_messages_updated_at 
BEFORE UPDATE ON broadcast_messages 
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();

-- Create trigger for resources updated_at
CREATE TRIGGER update_resources_updated_at 
BEFORE UPDATE ON resources 
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- DOWN
-- ============================================================================

-- Drop triggers
DROP TRIGGER IF EXISTS update_resources_updated_at ON resources;
DROP TRIGGER IF EXISTS update_broadcast_messages_updated_at ON broadcast_messages;

-- Drop views
DROP VIEW IF EXISTS user_assigned_tasks;
DROP VIEW IF EXISTS user_enrolled_projects;

-- Drop indexes
DROP INDEX IF EXISTS idx_resources_deleted_at;
DROP INDEX IF EXISTS idx_resources_download_count;
DROP INDEX IF EXISTS idx_resources_created_at;
DROP INDEX IF EXISTS idx_resources_tags;
DROP INDEX IF EXISTS idx_resources_resource_type;
DROP INDEX IF EXISTS idx_resources_uploaded_by;
DROP INDEX IF EXISTS idx_resources_project_id;
DROP INDEX IF EXISTS idx_resources_scope;

DROP INDEX IF EXISTS idx_broadcast_reads_broadcast_id;
DROP INDEX IF EXISTS idx_broadcast_reads_user_id;
DROP INDEX IF EXISTS idx_broadcast_messages_deleted_at;
DROP INDEX IF EXISTS idx_broadcast_messages_created_at;
DROP INDEX IF EXISTS idx_broadcast_messages_expires_at;
DROP INDEX IF EXISTS idx_broadcast_messages_priority;
DROP INDEX IF EXISTS idx_broadcast_messages_target_audience;
DROP INDEX IF EXISTS idx_broadcast_messages_author;

DROP INDEX IF EXISTS idx_project_messages_subject;
DROP INDEX IF EXISTS idx_project_messages_scope;
DROP INDEX IF EXISTS idx_project_messages_recipient_team;
DROP INDEX IF EXISTS idx_project_messages_recipient_user;

-- Drop tables
DROP TABLE IF EXISTS resources;
DROP TABLE IF EXISTS broadcast_reads;
DROP TABLE IF EXISTS broadcast_messages;

-- Remove columns from project_messages
ALTER TABLE project_messages DROP CONSTRAINT IF EXISTS check_message_recipient;
ALTER TABLE project_messages DROP CONSTRAINT IF EXISTS check_resource_scope;
ALTER TABLE project_messages DROP COLUMN IF EXISTS subject;
ALTER TABLE project_messages DROP COLUMN IF EXISTS message_scope;
ALTER TABLE project_messages DROP COLUMN IF EXISTS recipient_team_id;
ALTER TABLE project_messages DROP COLUMN IF EXISTS recipient_user_id;
