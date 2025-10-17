# 🚨 URGENT: Production Database Migration for Messaging Autocomplete

## Problem
The messaging autocomplete feature is failing with **500 errors** because the production database is missing the new columns for universal messaging.

**Error:** `pq: column "recipient_user_id" of relation "project_messages" does not exist`

## Solution: Run This Migration (2 minutes)

### Option 1: Cloud Console Query Editor (Easiest)

1. **Go to:** https://console.cloud.google.com/sql/instances/civicweave-postgres?project=civicweave-474622

2. **Click:** "QUERY" button

3. **Copy and paste this SQL:**

```sql
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

-- Verify migration
SELECT 'Migration completed successfully!' as status;
```

4. **Press Enter** and wait for confirmation

5. **Refresh your browser** - messaging autocomplete will work!

---

## After Migration

Once the migration completes:
- ✅ Message recipient autocomplete will work
- ✅ Users can search for recipients
- ✅ Messages can be sent to users and projects
- ✅ No more 500 errors

## Verification

After running the migration, check these columns exist:
```sql
\d project_messages
```

Should see the new columns:
- recipient_user_id
- recipient_team_id
- subject
- message_scope
- task_id

---

**Status:** Ready to run the migration in Cloud Console
