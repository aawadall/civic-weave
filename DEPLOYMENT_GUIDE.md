# 🚀 Task Management Deployment Guide

## Quick Deployment Steps

Since you're logged into gcloud, here's the fastest way to deploy:

### 1. 🗄️ Deploy Database Changes

**Option A: Google Cloud Console (Recommended)**
1. Go to: https://console.cloud.google.com/sql/instances/civicweave-postgres/overview
2. Click "Open Cloud Shell" (terminal icon in top right)
3. Run: `gcloud sql connect civicweave-postgres --user=civicweave --database=civicweave`
4. Copy and paste the migration SQL (see below)

**Option B: Local with psql**
```bash
# Install psql (if you have sudo access)
sudo apt-get update && sudo apt-get install -y postgresql-client

# Connect and run migration
gcloud sql connect civicweave-postgres --user=civicweave --database=civicweave
```

### 2. 📋 Migration SQL to Run

Copy and paste this SQL into your database connection:

```sql
-- Task Management Enhancements Migration
-- Adds task comments, time logging, new statuses, and task-related messaging

-- Create task_comments table
CREATE TABLE task_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES project_tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    comment_text TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    edited_at TIMESTAMP
);

-- Create task_time_logs table
CREATE TABLE task_time_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES project_tasks(id) ON DELETE CASCADE,
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    hours DECIMAL(5,2) NOT NULL CHECK (hours > 0),
    log_date DATE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add new status values to project_tasks
ALTER TABLE project_tasks 
DROP CONSTRAINT IF EXISTS project_tasks_status_check;

ALTER TABLE project_tasks 
ADD CONSTRAINT project_tasks_status_check 
CHECK (status IN ('todo', 'in_progress', 'done', 'blocked', 'takeover_requested'));

-- Add task_id and message_type to project_messages
ALTER TABLE project_messages 
ADD COLUMN IF NOT EXISTS task_id UUID REFERENCES project_tasks(id) ON DELETE CASCADE;

ALTER TABLE project_messages 
ADD COLUMN IF NOT EXISTS message_type VARCHAR(50) DEFAULT 'general' 
CHECK (message_type IN ('general', 'task_done', 'task_blocked', 'task_takeover'));

-- Create indexes
CREATE INDEX idx_task_comments_task_id ON task_comments(task_id);
CREATE INDEX idx_task_comments_user_id ON task_comments(user_id);
CREATE INDEX idx_task_comments_created_at ON task_comments(created_at);
CREATE INDEX idx_task_time_logs_task_id ON task_time_logs(task_id);
CREATE INDEX idx_task_time_logs_volunteer_id ON task_time_logs(volunteer_id);
CREATE INDEX idx_task_time_logs_log_date ON task_time_logs(log_date);
CREATE INDEX idx_project_messages_task_id ON project_messages(task_id);
CREATE INDEX idx_project_messages_type ON project_messages(message_type);

-- Helper functions for time aggregation
CREATE OR REPLACE FUNCTION get_task_total_hours(task_uuid UUID)
RETURNS DECIMAL(5,2) AS $$
BEGIN
    RETURN COALESCE(
        (SELECT SUM(hours) FROM task_time_logs WHERE task_id = task_uuid),
        0.00
    );
END;
$$ LANGUAGE plpgsql;

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

-- Create trigger for task_comments updated_at
CREATE TRIGGER update_task_comments_updated_at 
BEFORE UPDATE ON task_comments 
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();
```

### 3. 🚀 Deploy Application Services

After the database migration is complete:

```bash
# Build and push to production
make build-push

# Deploy to Cloud Run
make deploy-app
```

### 4. ✅ Verify Deployment

1. **Check Database**: Verify new tables exist
2. **Test Application**: Visit your deployed app
3. **Test Features**: 
   - Create a task as Team Lead
   - Self-assign as member
   - Add comments and log time
   - Change task status

## 🎯 What's New

After deployment, you'll have:

- **Task Comments**: Rich commenting system with progress updates
- **Time Logging**: Volunteers can log hours with automatic tallying
- **Status Transitions**: Mark tasks as done/blocked/request takeover
- **Auto-Notifications**: Task changes create TL notifications
- **Time Aggregation**: Automatic calculation of hours per task/volunteer/project

## 🔧 Troubleshooting

**Database Connection Issues:**
```bash
# Check gcloud authentication
gcloud auth list

# Check SQL instance status
gcloud sql instances list

# Test connection
gcloud sql connect civicweave-postgres --user=civicweave --database=civicweave
```

**Migration Errors:**
- Check if tables already exist
- Verify foreign key constraints
- Check for syntax errors in SQL

**Application Deployment Issues:**
```bash
# Check build status
make build-push

# Check Cloud Run status
gcloud run services list

# View logs
gcloud run services logs civicweave-backend --region=us-central1
```

## 📞 Support

If you encounter issues:
1. Check the error messages carefully
2. Verify database connection
3. Check Cloud Run service status
4. Review the PR: https://github.com/aawadall/civic-weave/pull/13

---

**Ready to deploy!** 🚀
