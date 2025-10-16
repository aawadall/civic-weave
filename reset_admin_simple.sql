-- Reset Admin Password to Simple Password
-- Password: admin123
-- Run this in Cloud SQL Console for PRODUCTION

UPDATE users 
SET password_hash = '$2b$12$t6DinqY/IpzxcKDLdpj94.7K6pGHL1T1qKcy7EUTP2kjnlGhrBQya'
WHERE email = 'admin@civicweave.com';

-- Ensure admin role is assigned
INSERT INTO user_roles (user_id, role_id, assigned_at)
SELECT u.id, r.id, NOW()
FROM users u, roles r
WHERE u.email = 'admin@civicweave.com' AND r.name = 'admin'
ON CONFLICT DO NOTHING;

-- Verify
SELECT 
    u.email,
    string_agg(r.name, ', ') as roles,
    'Login with: admin@civicweave.com / admin123' as credentials
FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
WHERE u.email = 'admin@civicweave.com'
GROUP BY u.email;



