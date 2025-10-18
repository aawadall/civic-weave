-- Fix Production Schema - Add Missing Columns
-- This script adds the missing content_json column to the projects table

-- Add content_json column to projects table if it doesn't exist
ALTER TABLE projects ADD COLUMN IF NOT EXISTS content_json JSONB;

-- Add comment for documentation
COMMENT ON COLUMN projects.content_json IS 'Rich text editor JSON content (TipTap format)';

-- Verify the column was added
SELECT column_name, data_type, is_nullable 
FROM information_schema.columns 
WHERE table_name = 'projects' AND column_name = 'content_json';
