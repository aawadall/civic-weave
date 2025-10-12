# Immediate Actions Required

## 🚨 Current Status Summary

### ✅ **What's Working:**
1. Backend deployed and running
2. Frontend deployed and running
3. Database connected (no more connection errors)
4. Rollback protection working
5. Email feature flag working (disabled)

### ❌ **What's Broken:**
1. **Volunteer profile creation failing** - causing registration to fail
2. **Likely cause**: Database migrations not run on production

### 🔐 **Security Issues Found:**
1. Database was exposed to internet (`0.0.0.0/0`) - **FIXED** ✅
2. Admin setup endpoint is public - **NEEDS FIX** ⚠️
3. Skill creation endpoint is public - **NEEDS FIX** ⚠️

## 📋 Action Plan (In Order)

### 1. Run Production Migrations (URGENT)

The volunteer table likely doesn't exist or is missing columns.

**Option A: Via Cloud SQL Proxy (Recommended)**
```bash
# Install Cloud SQL Proxy
curl -o cloud-sql-proxy https://storage.googleapis.com/cloud-sql-connectors/cloud-sql-proxy/v2.8.0/cloud-sql-proxy.linux.amd64
chmod +x cloud-sql-proxy

# Start proxy in background
./cloud-sql-proxy civicweave-474622:us-central1:civicweave-postgres &

# Run migrations
cd backend
export DB_HOST=127.0.0.1
export DB_PORT=5432
export DB_NAME=civicweave
export DB_USER=civicweave
export DB_PASSWORD=$(gcloud secrets versions access latest --secret="db-password")
go run cmd/migrate/main.go up

# Stop proxy
pkill cloud-sql-proxy
```

**Option B: Via Cloud Run Job (Better for production)**
```bash
# Create a migration job container
# Deploy as Cloud Run Job
# Run once to apply migrations
```

**Option C: Via Backend on Startup (Easiest)**
Already implemented! The backend runs migrations on startup (line 36 in main.go).
Check logs to see if it succeeded.

### 2. Seed Admin User

```bash
# Get the admin password
ADMIN_PASSWORD=$(gcloud secrets versions access latest --secret="admin-password")

# Run seed command
cd backend
export DB_HOST=127.0.0.1  # via Cloud SQL Proxy
export DB_PORT=5432
export DB_NAME=civicweave
export DB_USER=civicweave
export DB_PASSWORD=$(gcloud secrets versions access latest --secret="db-password")
export ADMIN_EMAIL=admin@civicweave.com
export ADMIN_PASSWORD=$ADMIN_PASSWORD
export ADMIN_NAME="System Administrator"
go run cmd/seed/main.go
```

### 3. Security Fixes (After Registration Works)

#### 3a. Protect Admin Setup
```bash
# Disable /api/admin/setup after first admin exists
# OR require SETUP_SECRET environment variable
```

#### 3b. Protect Skill Creation
```bash
# Move POST /api/skills/taxonomy to protected routes
# Require authentication to add skills
```

#### 3c. Update CORS (restrict origins)
```bash
# Add ALLOWED_ORIGINS env var
# Restrict to production frontend domain
```

### 4. Deploy Security Fixes
```bash
make build-push
make deploy-app
```

## 🧪 Quick Verification Commands

### Check if migrations ran:
```bash
gcloud run services logs read civicweave-backend --region=us-central1 | grep -i migration
```

### Check tables exist:
```bash
# Via Cloud SQL Proxy
psql "host=127.0.0.1 port=5432 dbname=civicweave user=civicweave password=XXX"
\dt
```

### Test registration again:
```bash
curl -X POST https://civicweave-backend-162941711179.us-central1.run.app/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"newuser@test.com","password":"password123","name":"New User","consent_given":true}'
```

## 📊 Current Deployment Info

- **Backend**: `civicweave-backend-00034-sgp`
- **Frontend**: `civicweave-frontend-00020-bfw` 
- **Database**: Private IP only ✅
- **Email**: Disabled (ENABLE_EMAIL=false) ✅

## 🎯 Expected Outcome

After running migrations:
1. Registration completes successfully
2. Users can login
3. All endpoints work as expected

Then after security fixes:
4. Admin setup endpoint protected
5. Skill creation requires auth
6. CORS restricted to known origins

## ⏱️ Time Estimate

- Migrations: 5 minutes
- Admin seed: 2 minutes
- Security fixes: 15 minutes
- Testing: 10 minutes
- **Total: ~30 minutes**

