-- Migration 007: Fix project created_by_admin_id foreign key to reference users table
-- This fixes the foreign key constraint that was referencing the old admins table

-- +migrate Up
-- Drop the old foreign key constraint that references admins table
ALTER TABLE projects 
    DROP CONSTRAINT IF EXISTS initiatives_created_by_admin_id_fkey;

-- Add new foreign key constraint that references users table
ALTER TABLE projects 
    ADD CONSTRAINT projects_created_by_admin_id_fkey 
    FOREIGN KEY (created_by_admin_id) 
    REFERENCES users(id) 
    ON DELETE CASCADE;

-- +migrate Down
-- Revert to the old constraint (though this would fail if admins table doesn't exist)
ALTER TABLE projects 
    DROP CONSTRAINT IF EXISTS projects_created_by_admin_id_fkey;

ALTER TABLE projects 
    ADD CONSTRAINT initiatives_created_by_admin_id_fkey 
    FOREIGN KEY (created_by_admin_id) 
    REFERENCES admins(id);

