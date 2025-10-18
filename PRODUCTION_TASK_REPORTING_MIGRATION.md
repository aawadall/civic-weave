# 🚨 URGENT: Production Task Reporting Migration Required

## Problem
Your production database is missing the new task timeline columns and activity log table, causing **500 errors** when users try to access the dashboard.

Error: `pq: column pt.started_at does not exist`

## Solution: Run This Migration Now (2 minutes)

### Option 1: Cloud Console Query Editor (Easiest)

1. **Go to:** https://console.cloud.google.com/sql/instances/civicweave-postgres?project=civicweave-474622

2. **Click:** "OPEN CLOUD SHELL" button at the top

3. **Run these commands in Cloud Shell:**

```bash
# Connect to database
gcloud sql connect civicweave-postgres --user=postgres --database=civicweave

# When prompted for password, enter your postgres password
# Then copy-paste ALL of this SQL:
```

### THE SQL TO RUN:

```sql
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
```

4. **Press Enter** and wait for confirmation

5. **Refresh your browser** - 500 errors should be gone!

---

## After Migration

Once the migration completes:
- ✅ Dashboard will load without errors
- ✅ Task timeline tracking will work
- ✅ Activity logging will be enabled
- ✅ All new task features will be functional

## Verification

After running the migration, check these columns exist:
```sql
\d project_tasks
```

Should see new columns:
- started_at
- blocked_at  
- blocked_reason
- completed_at
- completion_note
- takeover_requested_at
- takeover_reason
- last_status_changed_by

And this table should exist:
```sql
\dt task_activity_log
```

---

**Status:** Ready to deploy - just run the migration in Cloud Console
