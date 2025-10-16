-- EMERGENCY FIX: Drop role column that migration 010 failed to drop
-- Run this in Cloud SQL Console immediately

-- Step 1: Drop the role column (the real issue)
ALTER TABLE users DROP COLUMN IF EXISTS role;

-- Step 2: Verify it's gone
SELECT 'Role column exists:' as check, 
       EXISTS (
           SELECT 1 FROM information_schema.columns 
           WHERE table_name = 'users' AND column_name = 'role'
       ) as result;

-- Step 3: Also ensure roles exist (in case they weren't created)
INSERT INTO roles (id, name, description, permissions, created_at) VALUES
('11111111-1111-1111-1111-111111111111', 'admin', 'Administrator', '["*"]'::jsonb, NOW()),
('22222222-2222-2222-2222-222222222222', 'volunteer', 'Volunteer', '["view"]'::jsonb, NOW()),
('33333333-3333-3333-3333-333333333333', 'team_lead', 'Team Lead', '["manage"]'::jsonb, NOW()),
('44444444-4444-4444-4444-444444444444', 'campaign_manager', 'Campaign Manager', '["campaign"]'::jsonb, NOW())
ON CONFLICT (name) DO NOTHING;

-- Step 4: Verify roles
SELECT 'Roles count:' as check, COUNT(*) as result FROM roles;

-- Done! Now test registration immediately - no restart needed



