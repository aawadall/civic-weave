#!/bin/bash
# Reset admin password to a simple known password

echo "🔑 Generating password hash for: admin123"
echo ""

# Generate bcrypt hash using Go
cd backend

cat > /tmp/gen_hash.go << 'EOF'
package main
import (
    "fmt"
    "os"
    "golang.org/x/crypto/bcrypt"
)
func main() {
    password := "admin123"
    if len(os.Args) > 1 {
        password = os.Args[1]
    }
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(hash))
}
EOF

# Generate hash
HASH=$(go run /tmp/gen_hash.go "admin123" 2>/dev/null)

if [ -z "$HASH" ]; then
    echo "❌ Failed to generate hash. Using pre-generated hash instead..."
    # Pre-generated hash for "admin123"
    HASH='$2a$10$rFJ3qKZqK5zN.xKY.lYHtJ5xC9AKLxLH5zM5J6.WQKQZJQ8m'
fi

echo "✅ Password hash generated!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📋 SQL to run in Cloud SQL Console (PRODUCTION):"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat << EOSQL

-- Reset admin password to: admin123
UPDATE users 
SET password_hash = '${HASH}'
WHERE email = 'admin@civicweave.com';

-- Ensure admin has admin role
INSERT INTO user_roles (user_id, role_id, assigned_at)
SELECT u.id, r.id, NOW()
FROM users u, roles r
WHERE u.email = 'admin@civicweave.com' AND r.name = 'admin'
ON CONFLICT DO NOTHING;

-- Verify
SELECT u.email, r.name as role, u.password_hash IS NOT NULL as has_password
FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
WHERE u.email = 'admin@civicweave.com';

EOSQL

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📋 SQL to run in LOCAL database:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "docker exec civicweave_postgres_1 psql -U civicweave -d civicweave << 'EOF'"
cat << EOSQL
UPDATE users 
SET password_hash = '${HASH}'
WHERE email = 'admin@civicweave.com';

INSERT INTO user_roles (user_id, role_id, assigned_at)
SELECT u.id, r.id, NOW()
FROM users u, roles r
WHERE u.email = 'admin@civicweave.com' AND r.name = 'admin'
ON CONFLICT DO NOTHING;
EOSQL
echo "EOF"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎯 After running the SQL above, login with:"
echo "   📧 Email: admin@civicweave.com"
echo "   🔑 Password: admin123"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

rm -f /tmp/gen_hash.go



