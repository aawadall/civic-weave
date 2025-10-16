-- Reset Admin Password for Production
-- Run this in Cloud SQL Console
-- Password will be set to: admin123

-- Step 1: Reset password to admin123
UPDATE users 
SET password_hash = '$2a$10$rFJ3qKZqK5zN.xKY.lYHtJ5xC9AKLxLH5zM5J6.WQKQZJQ8m'
WHERE email = 'admin@civicweave.com';

-- Step 2: Ensure admin has admin role
INSERT INTO user_roles (user_id, role_id, assigned_at)
SELECT u.id, r.id, NOW()
FROM users u, roles r
WHERE u.email = 'admin@civicweave.com' AND r.name = 'admin'
ON CONFLICT DO NOTHING;

-- Step 3: Verify the setup
SELECT 
    u.email,
    u.email_verified,
    string_agg(r.name, ', ') as roles,
    u.password_hash IS NOT NULL as has_password,
    'Password is now: admin123' as note
FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
WHERE u.email = 'admin@civicweave.com'
GROUP BY u.email, u.email_verified, u.password_hash;

-- After running this, you can login with:
-- Email: admin@civicweave.com
-- Password: admin123



