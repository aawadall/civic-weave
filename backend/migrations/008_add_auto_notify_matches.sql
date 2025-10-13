-- UP
-- Add auto_notify_matches field to projects table
-- This field enables "hunt mode" where top-matching volunteers are automatically notified

ALTER TABLE projects ADD COLUMN IF NOT EXISTS auto_notify_matches BOOLEAN DEFAULT false;

COMMENT ON COLUMN projects.auto_notify_matches IS 'When enabled, automatically notifies top-matching volunteers when batch job runs. Used for urgent or hard-to-fill projects.';

-- DOWN
ALTER TABLE projects DROP COLUMN IF EXISTS auto_notify_matches;

