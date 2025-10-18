#!/bin/bash
# Script to run task reporting migration on production Cloud SQL

echo "🚀 Running Task Reporting Migration on Production Cloud SQL..."
echo ""
echo "This will:"
echo "  1. Add timeline columns to project_tasks table"
echo "  2. Create task_activity_log table"
echo "  3. Back-fill activity log with existing tasks"
echo "  4. Drop legacy project columns"
echo ""

# Extract the SQL from our v2 migration
cat > /tmp/task_reporting_migration.sql << 'EOF'
-- ============================================================================
-- TASK REPORTING AND TIMELINE TRACKING MIGRATION
-- ============================================================================

-- PART 1: Add Timeline Columns to project_tasks
ALTER TABLE project_tasks 
ADD COLUMN IF NOT EXISTS started_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS blocked_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS blocked_reason TEXT,
ADD COLUMN IF NOT EXISTS completed_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS completion_note TEXT,
ADD COLUMN IF NOT EXISTS takeover_requested_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS takeover_reason TEXT,
ADD COLUMN IF NOT EXISTS last_status_changed_by UUID REFERENCES users(id);

-- PART 2: Create task_activity_log table
CREATE TABLE IF NOT EXISTS task_activity_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES project_tasks(id) ON DELETE CASCADE,
    actor_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_volunteer_id UUID REFERENCES volunteers(id) ON DELETE CASCADE,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    context JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- PART 3: Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_task_activity_log_task_id ON task_activity_log(task_id);
CREATE INDEX IF NOT EXISTS idx_task_activity_log_created_at ON task_activity_log(created_at);
CREATE INDEX IF NOT EXISTS idx_task_activity_log_actor_user_id ON task_activity_log(actor_user_id);

-- PART 4: Back-fill task_activity_log with existing data
INSERT INTO task_activity_log (task_id, actor_user_id, from_status, to_status, context, created_at)
SELECT 
    pt.id as task_id,
    pt.created_by_id as actor_user_id,
    NULL as from_status,
    pt.status as to_status,
    '{"initial_creation": true}'::jsonb as context,
    pt.created_at
FROM project_tasks pt
WHERE pt.created_at IS NOT NULL
ON CONFLICT (id) DO NOTHING;

-- PART 5: Drop legacy columns from projects table (if they exist)
DROP INDEX IF EXISTS idx_projects_required_skills;
DROP INDEX IF EXISTS idx_projects_status;
ALTER TABLE projects DROP COLUMN IF EXISTS required_skills;
ALTER TABLE projects DROP COLUMN IF EXISTS status;

-- PART 6: Ensure updated_at trigger exists for project_tasks
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

-- PART 7: Verify migration
SELECT 'Task reporting migration completed successfully!' as status;
SELECT COUNT(*) as project_tasks_count FROM project_tasks;
SELECT COUNT(*) as task_activity_log_count FROM task_activity_log;
EOF

echo "📁 Migration file prepared at /tmp/task_reporting_migration.sql"
echo ""
echo "Choose your method:"
echo ""
echo "Option 1: Use gcloud CLI (requires psql installed)"
echo "  Run: gcloud sql connect civicweave-postgres --user=postgres --database=civicweave"
echo "  Then: \\i /tmp/task_reporting_migration.sql"
echo ""
echo "Option 2: Cloud Console Query Editor"
echo "  1. Go to: https://console.cloud.google.com/sql/instances/civicweave-postgres?project=civicweave-474622"
echo "  2. Click 'QUERY' button"
echo "  3. Copy and paste the SQL below:"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat /tmp/task_reporting_migration.sql
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "After running the migration, the 500 errors will be fixed!"
echo "The dashboard will load properly with timeline tracking enabled."
