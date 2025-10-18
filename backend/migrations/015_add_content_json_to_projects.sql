-- Migration: Add content_json column to projects table
-- Date: 2025-01-XX
-- Description: Support rich text project descriptions

-- Up Migration
ALTER TABLE projects 
ADD COLUMN IF NOT EXISTS content_json JSONB;

-- Add comment for documentation
COMMENT ON COLUMN projects.content_json IS 'Rich text editor JSON content (TipTap format)';

-- No index needed initially - this is read-only display field
