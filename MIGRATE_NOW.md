# 🚨 CRITICAL: Run This Command NOW to Fix 500 Errors

## The Problem
Your production database is missing new tables. The backend is trying to query them and failing with 500 errors.

## The Solution (30 seconds)

**Copy and paste this ONE command in your terminal:**

```bash
gcloud sql connect civicweave-postgres --user=postgres --database=civicweave < /tmp/prod_migration_final.sql
```

**When prompted for password:**
- Enter your postgres password (the one from your secrets)

**That's it!** The migration will run automatically.

---

## If That Doesn't Work

**Alternative - Use Cloud Console (2 min):**

1. Open: https://console.cloud.google.com/sql/instances/civicweave-postgres/query?project=civicweave-474622

2. Copy EVERYTHING between the ═══ lines below and paste in the query box:

═══════════════════════════════════════════════════════════════════

```sql
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

ALTER TABLE projects ADD COLUMN IF NOT EXISTS content_json JSONB DEFAULT NULL;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS budget_total DECIMAL(12, 2) DEFAULT 0.00;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS budget_spent DECIMAL(12, 2) DEFAULT 0.00;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS permissions JSONB DEFAULT '{}';

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

CREATE TRIGGER update_project_tasks_updated_at 
BEFORE UPDATE ON project_tasks 
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();
```

═══════════════════════════════════════════════════════════════════

3. Click "RUN"

4. Done!

---

## After Migration

Your site will immediately work:
- ✅ Projects page loads (no 500 error)
- ✅ Tasks tab functional
- ✅ Messages tab functional
- ✅ Logistics tab functional

---

**Run one of these methods NOW to fix the 500 errors!**

