#!/bin/bash

# Get database password from secret manager
DB_PASSWORD=$(gcloud secrets versions access latest --secret="db-password")

echo "Creating admin user directly in database..."

# Create a simple admin user creation script
cat > /tmp/create_admin.sql << EOF
-- Create admin user directly
INSERT INTO users (id, email, password_hash, email_verified, role, created_at, updated_at) 
VALUES (
    gen_random_uuid(),
    'admin@civicweave.com', 
    '\$2a\$10\$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    true, 
    'admin', 
    NOW(), 
    NOW()
) ON CONFLICT (email) DO UPDATE SET 
    password_hash = EXCLUDED.password_hash,
    email_verified = true,
    role = 'admin',
    updated_at = NOW();

-- Create admin profile
INSERT INTO admins (id, user_id, name, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    u.id,
    'System Administrator',
    NOW(),
    NOW()
FROM users u 
WHERE u.email = 'admin@civicweave.com'
ON CONFLICT (user_id) DO UPDATE SET
    name = EXCLUDED.name,
    updated_at = NOW();

SELECT 'Admin user created successfully' as result;
EOF

# Execute the SQL (we'll need to use a different approach since psql is not available)
echo "SQL script created. You can run this manually in your database:"
cat /tmp/create_admin.sql
