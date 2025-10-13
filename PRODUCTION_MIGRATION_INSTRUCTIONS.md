# 🚨 URGENT: Production Database Migration Required

## Problem
Your production database is missing the new tables, causing **500 errors** when users try to:
- View projects
- Access tasks
- Send messages
- Use logistics features

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
-- Project Tasks System
CREATE TABLE IF NOT EXISTS project_tasks (
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

CREATE TABLE IF NOT EXISTS task_updates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES project_tasks(id) ON DELETE CASCADE,
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    update_text TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_project_tasks_project_id ON project_tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_project_tasks_assignee_id ON project_tasks(assignee_id);
CREATE INDEX IF NOT EXISTS idx_project_tasks_status ON project_tasks(status);
CREATE INDEX IF NOT EXISTS idx_project_tasks_due_date ON project_tasks(due_date);
CREATE INDEX IF NOT EXISTS idx_project_tasks_created_by ON project_tasks(created_by_id);
CREATE INDEX IF NOT EXISTS idx_task_updates_task_id ON task_updates(task_id);
CREATE INDEX IF NOT EXISTS idx_task_updates_volunteer_id ON task_updates(volunteer_id);

-- Messaging System
CREATE TABLE IF NOT EXISTS project_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message_text TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    edited_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS message_reads (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message_id UUID NOT NULL REFERENCES project_messages(id) ON DELETE CASCADE,
    read_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, message_id)
);

CREATE INDEX IF NOT EXISTS idx_project_messages_project_id ON project_messages(project_id);
CREATE INDEX IF NOT EXISTS idx_project_messages_sender_id ON project_messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_project_messages_created_at ON project_messages(created_at);
CREATE INDEX IF NOT EXISTS idx_project_messages_deleted_at ON project_messages(deleted_at);
CREATE INDEX IF NOT EXISTS idx_message_reads_user_id ON message_reads(user_id);
CREATE INDEX IF NOT EXISTS idx_message_reads_message_id ON message_reads(message_id);

-- Enhanced Projects Schema
ALTER TABLE projects ADD COLUMN IF NOT EXISTS content_json JSONB DEFAULT NULL;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS budget_total DECIMAL(12, 2) DEFAULT 0.00;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS budget_spent DECIMAL(12, 2) DEFAULT 0.00;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS permissions JSONB DEFAULT '{}';

COMMENT ON COLUMN projects.content_json IS 'TipTap editor JSON format for rich text project description';

-- Application Auto-Enrollment Trigger
CREATE OR REPLACE FUNCTION auto_enroll_volunteer()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'accepted' AND (OLD.status IS NULL OR OLD.status != 'accepted') THEN
        INSERT INTO project_team_members (project_id, volunteer_id, status, joined_at)
        VALUES (NEW.project_id, NEW.volunteer_id, 'active', CURRENT_TIMESTAMP)
        ON CONFLICT (project_id, volunteer_id) DO UPDATE
        SET status = 'active', updated_at = CURRENT_TIMESTAMP;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_auto_enroll_volunteer ON applications;
CREATE TRIGGER trigger_auto_enroll_volunteer
AFTER UPDATE ON applications
FOR EACH ROW
WHEN (NEW.status = 'accepted')
EXECUTE FUNCTION auto_enroll_volunteer();

-- Timestamp Trigger
CREATE TRIGGER update_project_tasks_updated_at 
BEFORE UPDATE ON project_tasks 
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();

-- Verify migration
SELECT 'Migration completed successfully!' as status;
SELECT COUNT(*) as project_tasks_count FROM project_tasks;
SELECT COUNT(*) as project_messages_count FROM project_messages;
```

4. **Press Enter** and wait for confirmation

5. **Refresh your browser** - 500 errors should be gone!

---

### Option 2: If Cloud Shell Doesn't Work

Save this to a file and use Cloud SQL Proxy, or contact me to help troubleshoot.

---

## After Migration

Once the migration completes:
- ✅ Projects page will load
- ✅ All tabs will work
- ✅ No more 500 errors
- ✅ Tasks, messages, logistics fully functional

## Verification

After running the migration, check these tables exist:
```sql
\dt project_*
\dt message_*
\dt task_*
```

Should see:
- project_tasks
- project_messages  
- message_reads
- task_updates

---

**Status:** Waiting for you to run the migration in Cloud Console

