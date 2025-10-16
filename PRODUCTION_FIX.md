# Production Fix Summary

## ✅ What's Fixed
1. **Many-to-Many Roles Migration** - Successfully applied to production
2. **Backend Code Deployed** - Latest code with many-to-many roles support
3. **Database Connected** - Cloud Run properly connected to Cloud SQL
4. **Idempotent Migrations** - All migrations can run safely multiple times

## ❌ Current Issue
**Root Cause**: Production database is missing default roles (volunteer, admin, team_lead, campaign_manager)

**Symptoms**:
- User registration fails with "Failed to create user"
- Login fails with "Invalid credentials" (password mismatch + no roles)
- Users exist but have no roles assigned

## 🚀 Solution Required

### Step 1: Insert Default Roles

Run this SQL in production database:

```sql
-- Insert default roles
INSERT INTO roles (id, name, description, permissions, created_at) VALUES
(gen_random_uuid(), 'volunteer', 'Can view and apply to projects, manage own profile', '["view_projects", "apply_to_projects", "view_own_profile", "rate_volunteers"]', CURRENT_TIMESTAMP),
(gen_random_uuid(), 'team_lead', 'Can create/edit projects, manage teams, rate volunteers', '["view_projects", "create_projects", "edit_projects", "manage_teams", "rate_volunteers", "view_volunteers"]', CURRENT_TIMESTAMP),
(gen_random_uuid(), 'campaign_manager', 'Can create and send email campaigns', '["view_projects", "view_volunteers", "create_campaigns", "send_campaigns", "view_campaigns"]', CURRENT_TIMESTAMP),
(gen_random_uuid(), 'admin', 'Full system access including user and role management', '["*"]', CURRENT_TIMESTAMP)
ON CONFLICT (name) DO NOTHING;
```

### Step 2: Assign Admin Role to Existing Admin User

```sql
-- Assign admin role to admin@civicweave.com
INSERT INTO user_roles (user_id, role_id, assigned_at)
SELECT 
    u.id,
    r.id,
    CURRENT_TIMESTAMP
FROM users u
CROSS JOIN roles r
WHERE u.email = 'admin@civicweave.com' 
  AND r.name = 'admin'
ON CONFLICT (user_id, role_id) DO NOTHING;
```

### Step 3: Reset Admin Password (Optional)

If password is unknown, generate new hash and update:

```bash
# Generate hash locally
cd backend
docker run --rm golang:1.23-alpine sh -c '
  apk add --no-cache git && \
  echo "package main
import (\"fmt\"; \"golang.org/x/crypto/bcrypt\")
func main() {
  hash, _ := bcrypt.GenerateFromPassword([]byte(\"admin4civicweave\"), 10)
  fmt.Println(string(hash))
}" > /tmp/main.go && \
  cd /tmp && go mod init tmp && go get golang.org/x/crypto/bcrypt && go run main.go
'
```

Then update password:
```sql
UPDATE users 
SET password_hash = '<generated_hash>'
WHERE email = 'admin@civicweave.com';
```

## Alternative: Use Cloud Console SQL Editor

1. Go to: https://console.cloud.google.com/sql/instances/civicweave-postgres/query?project=civicweave-474622
2. Run the SQL commands above
3. Test login

## Verification

After fixing:
```bash
# Test registration
curl -X POST https://civicweave-backend-162941711179.us-central1.run.app/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"Test1234!","name":"Test","consent_given":true}'

# Test login
curl -X POST https://civicweave-backend-162941711179.us-central1.run.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@civicweave.com","password":"admin4civicweave"}'
```

## System Status

✅ **Code**: Fully migrated to many-to-many roles
✅ **Database Schema**: Migration 010 applied successfully  
✅ **Deployment**: Backend running with correct code
❌ **Data**: Missing default roles in production database



