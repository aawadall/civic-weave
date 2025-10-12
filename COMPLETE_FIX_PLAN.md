# Complete Fix Plan - Registration & Security

## 🐛 Root Cause Analysis

### Problem 1: Migration Failure
```
Error: constraint "project_team_members_project_id_temp_fkey" does not exist
```
**Cause**: Migration 004 tries to drop a constraint that may not exist  
**Impact**: Migrations stopped at step 3, tables from migrations 4-5 don't exist  
**Result**: Volunteer table might be missing columns or constraints

### Problem 2: Security Gaps
- Database was exposed to `0.0.0.0/0` (FIXED ✅)
- Admin setup endpoint is public (NEEDS FIX)
- Skill creation is public (NEEDS FIX)

## 🎯 Complete Solution (Step by Step)

### Step 1: Fix Migration 004 (Make it Idempotent)

Update the migration to handle missing constraints gracefully:

```sql
-- Instead of:
ALTER TABLE project_team_members DROP CONSTRAINT project_team_members_project_id_temp_fkey;

-- Use:
DO $$
BEGIN
    ALTER TABLE project_team_members DROP CONSTRAINT IF EXISTS project_team_members_project_id_temp_fkey;
EXCEPTION
    WHEN undefined_object THEN
        NULL; -- Constraint doesn't exist, ignore
END $$;
```

### Step 2: Reset Production Database (Safest if No Real Users Yet)

```bash
# Connect to Cloud SQL
gcloud sql connect civicweave-postgres --user=civicweave --database=civicweave

# In psql:
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO civicweave;
GRANT ALL ON SCHEMA public TO public;
\q
```

Then the backend will auto-run all migrations on next startup.

### Step 3: OR Manually Complete Migrations

```bash
# Run migrations via backend CLI
cd /home/arashad/src/CivicWeave/backend

# Set up Cloud SQL Proxy connection
export DB_HOST=/cloudsql/civicweave-474622:us-central1:civicweave-postgres
export DB_NAME=civicweave
export DB_USER=civicweave
export DB_PASSWORD=$(gcloud secrets versions access latest --secret="db-password")

# Run migration tool
go run cmd/migrate/main.go up
```

### Step 4: Seed Admin User

```bash
cd /home/arashad/src/CivicWeave/backend

export DB_HOST=/cloudsql/civicweave-474622:us-central1:civicweave-postgres
export DB_NAME=civicweave
export DB_USER=civicweave
export DB_PASSWORD=$(gcloud secrets versions access latest --secret="db-password")
export ADMIN_EMAIL=admin@civicweave.com
export ADMIN_PASSWORD=$(gcloud secrets versions access latest --secret="admin-password")
export ADMIN_NAME="System Administrator"

go run cmd/seed/main.go
```

### Step 5: Fix Security Issues in Code

#### 5a. Fix migration 004
```bash
# Edit backend/migrations/004_rename_initiatives_to_projects.sql
# Add IF EXISTS to all DROP CONSTRAINT commands
```

#### 5b. Protect admin setup
```bash
# Edit backend/cmd/server/main.go
# Add check: only allow if no admin exists
```

#### 5c. Protect skill creation
```bash
# Move POST /skills/taxonomy to protected routes
```

### Step 6: Deploy All Fixes

```bash
make build-push
make deploy-infra  # Apply DB network security
make deploy-app    # Deploy code fixes
```

## 🚀 Quick Fix Option (If No Production Data)

If there are NO real users in production yet:

```bash
# 1. Reset database completely
gcloud sql databases delete civicweave --instance=civicweave-postgres
gcloud sql databases create civicweave --instance=civicweave-postgres

# 2. Restart backend (will auto-run all migrations)
gcloud run services update civicweave-backend --region=us-central1

# 3. Wait 30 seconds, check logs
gcloud run services logs read civicweave-backend --region=us-central1 --limit=50

# 4. Seed admin
# (connect via Cloud SQL Proxy and run seed command)

# 5. Test registration
curl -X POST https://civicweave-backend-162941711179.us-central1.run.app/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"password123","name":"Test User","consent_given":true}'
```

## 📊 Decision Tree

```
Do you have real user data in production?
├─ NO  → Use Quick Fix (reset database)
│        ├─ Fix migration 004
│        ├─ Delete & recreate database
│        ├─ Migrations auto-run on startup
│        └─ Seed admin user
│
└─ YES → Use Manual Migration Fix
         ├─ Fix migration 004 to be idempotent
         ├─ Connect via Cloud SQL Proxy
         ├─ Run migrations manually
         ├─ Verify data integrity
         └─ Deploy fixed migration
```

## ✅ What to Do NOW

**My Recommendation**: Since this appears to be early development:

1. **Delete and recreate the database** (Quick Fix above)
2. **Fix migration 004** to prevent future issues
3. **Deploy all changes** including security fixes
4. **Test thoroughly** before adding real users

**Estimated Time**: 15 minutes total

Would you like me to:
- **A**: Create the fixed migration 004 and other security fixes?
- **B**: Provide commands to reset the database?
- **C**: Both A and B?

