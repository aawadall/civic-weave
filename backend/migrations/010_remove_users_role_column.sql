-- UP
-- Migrate from single role column to many-to-many user_roles relationship

-- Step 1: Migrate existing role data to user_roles table (only if role column exists)
-- For each user with an old 'role' value, assign them the corresponding role from the roles table
DO $$
BEGIN
    IF EXISTS (
        SELECT FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'role'
    ) THEN
        INSERT INTO user_roles (user_id, role_id, assigned_at)
        SELECT 
            u.id AS user_id,
            r.id AS role_id,
            u.created_at AS assigned_at
        FROM users u
        INNER JOIN roles r ON r.name = u.role
        WHERE u.role IS NOT NULL
        ON CONFLICT (user_id, role_id) DO NOTHING;
    END IF;
END $$;

-- Step 2: Drop the index on the old role column (if it exists)
DROP INDEX IF EXISTS idx_users_role;

-- Step 3: Drop the old role column from users table (if it exists)
ALTER TABLE users DROP COLUMN IF EXISTS role;

-- DOWN
-- Restore the role column (best effort - will use primary role if user has multiple)
ALTER TABLE users ADD COLUMN role VARCHAR(20) CHECK (role IN ('admin', 'volunteer', 'team_lead', 'campaign_manager'));

-- Recreate the index
CREATE INDEX idx_users_role ON users(role);

-- Migrate back from user_roles to single role column
-- Use the first role alphabetically if a user has multiple roles
UPDATE users u
SET role = (
    SELECT r.name
    FROM user_roles ur
    INNER JOIN roles r ON r.id = ur.role_id
    WHERE ur.user_id = u.id
    ORDER BY r.name
    LIMIT 1
);

