-- DOWN
-- Remove content_json column from projects table
-- This migration removes the content_json column

-- Remove content_json column
ALTER TABLE projects DROP COLUMN IF EXISTS content_json;
