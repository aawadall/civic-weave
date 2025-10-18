-- UP
-- Add content_json column to projects table
-- This migration adds the missing content_json column for rich text project descriptions

-- Add content_json column to projects table if it doesn't exist
ALTER TABLE projects ADD COLUMN IF NOT EXISTS content_json JSONB;

-- Add comment for documentation
COMMENT ON COLUMN projects.content_json IS 'Rich text editor JSON content (TipTap format)';
