-- Fix project created_by_admin_id foreign key to reference users table
-- Run this in your database console

-- Drop the old foreign key constraint that references admins table
ALTER TABLE projects 
    DROP CONSTRAINT IF EXISTS initiatives_created_by_admin_id_fkey;

-- Add new foreign key constraint that references users table
ALTER TABLE projects 
    ADD CONSTRAINT projects_created_by_admin_id_fkey 
    FOREIGN KEY (created_by_admin_id) 
    REFERENCES users(id) 
    ON DELETE CASCADE;

-- Verify the constraint was added
SELECT 
    conname AS constraint_name,
    conrelid::regclass AS table_name,
    confrelid::regclass AS references_table
FROM pg_constraint 
WHERE conname = 'projects_created_by_admin_id_fkey';

