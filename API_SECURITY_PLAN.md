# API Security Plan

## 🔒 Current Security Status

### ✅ **Already Protected (Good)**

#### Authentication-Protected Endpoints
All routes under `protected.Use(middleware.AuthRequired())` require JWT:
- `/api/me` - User profile
- `/api/volunteers/*` - All volunteer operations
- `/api/projects/*` - All project operations  
- `/api/applications/*` - All applications
- `/api/matching/*` - All matching endpoints
- `/api/admin/*` - Admin endpoints (double-protected with role checks)

#### Rate Limiting
- ✅ Login: Rate limited
- ✅ Registration: Rate limited
- ✅ OAuth: Rate limited

#### CORS Protection
- ✅ CORS middleware enabled
- ⚠️ May need to restrict allowed origins in production

### ⚠️ **Security Gaps (Need Fixing)**

| Endpoint | Issue | Risk Level | Fix Needed |
|----------|-------|------------|------------|
| `POST /api/admin/setup` | No authentication | 🔴 **CRITICAL** | Disable after first use OR require secret token |
| `POST /api/skills/taxonomy` | Public | 🟡 **MEDIUM** | Require authentication |
| `GET /api/skills/taxonomy` | Public | 🟢 **LOW** | OK for now (read-only) |
| `/health` | Public | 🟢 **OK** | Fine for monitoring |
| Backend Cloud Run | `allUsers` | 🟢 **OK** | Correct for web API |

## 🛡️ Recommended Security Enhancements

### 1. **Critical: Secure Admin Setup Endpoint**

```go
// Option A: Disable after first admin exists
if adminSetupHandler != nil {
    // Only allow if no admin exists
    router.POST("/api/admin/setup", 
        middleware.RequireNoExistingAdmin(adminService),
        adminSetupHandler.CreateAdmin)
}

// Option B: Require secret token
router.POST("/api/admin/setup", 
    middleware.RequireSetupToken(cfg.SetupSecret),
    adminSetupHandler.CreateAdmin)
```

**Action**: Implement Option A (safer, no token to leak)

### 2. **Medium: Protect Skill Creation**

```go
// Move POST to protected routes
protected.POST("/skills/taxonomy", skillHandler.AddSkill)
// OR require admin role
protected.POST("/skills/taxonomy", 
    middleware.RequireRole("admin"), 
    skillHandler.AddSkill)
```

**Action**: Require authentication for adding skills

### 3. **CORS Hardening**

```go
// In middleware/cors.go - restrict origins
AllowedOrigins: []string{
    "https://civicweave-frontend-162941711179.us-central1.run.app",
    "http://localhost:3000", // dev only
}
```

**Action**: Whitelist specific origins

### 4. **API Gateway (Optional - Advanced)**

Add Google Cloud API Gateway in front of Cloud Run:
- ✅ Built-in API key management
- ✅ Rate limiting at gateway level
- ✅ Request/response transformation
- ✅ API versioning
- ❌ Additional cost (~$3/million requests)

**Action**: Consider for v2.0

### 5. **Cloud Armor (Optional - DDoS Protection)**

Add Cloud Armor Web Application Firewall:
- ✅ DDoS protection
- ✅ IP allowlist/blocklist
- ✅ Rate limiting by IP
- ✅ Bot protection
- ❌ Additional cost (~$7/policy + $1/million requests)

**Action**: Consider for production scale

### 6. **Request Signing (Optional - Advanced)**

Implement request signatures:
- Frontend signs requests with shared secret
- Backend verifies signatures
- Prevents API abuse even with public endpoints

**Action**: Consider if facing abuse

## 📋 Immediate Action Items

### Priority 1: Critical Fixes (Do Now)

1. **Disable Admin Setup After First Use**
   ```bash
   # Create middleware/admin_setup_guard.go
   # Update cmd/server/main.go
   # Deploy
   ```

2. **Protect Skill Creation**
   ```bash
   # Move POST /skills/taxonomy to protected routes
   # Deploy
   ```

### Priority 2: Important Hardening (This Week)

3. **CORS Whitelist**
   ```bash
   # Update middleware/cors.go
   # Add ALLOWED_ORIGINS env var
   # Deploy
   ```

4. **Database Network Isolation** (Already Done! ✅)
   - Disabled public IP
   - Using Cloud SQL Proxy
   - Unix socket connection

### Priority 3: Future Enhancements (Next Sprint)

5. **API Rate Limiting Improvements**
   - Per-user rate limits (not just global)
   - Different limits for different endpoints
   - Redis-backed distributed rate limiting

6. **Audit Logging**
   - Log all sensitive operations
   - Track failed auth attempts
   - Monitor for abuse patterns

7. **Security Headers**
   - Add security headers middleware
   - HSTS, CSP, X-Frame-Options, etc.

## 🔐 Production Security Checklist

### Before Going Live:

- [ ] **Disable admin setup endpoint** (or protect with token)
- [ ] **Protect skill creation endpoint**
- [ ] **Restrict CORS to specific origins**
- [ ] **Enable Cloud SQL deletion protection**
- [ ] **Rotate all secrets** (use production values)
- [ ] **Enable email verification** (set ENABLE_EMAIL=true)
- [ ] **Set up monitoring and alerts**
- [ ] **Review all public endpoints**
- [ ] **Test rate limiting thresholds**
- [ ] **Enable Cloud SQL automatic backups** (already enabled ✅)

### Network Security Status:

- ✅ **Cloud SQL**: Private IP only (no public access)
- ✅ **Redis**: Private IP only (GCP internal)
- ✅ **Secret Manager**: IAM-based access
- ⚠️ **Cloud Run Backend**: Public (must be for web API)
- ⚠️ **Cloud Run Frontend**: Public (must be for web app)

### Authentication Status:

| Layer | Status | Details |
|-------|--------|---------|
| **Database** | ✅ Password + Private Network | Cloud SQL Proxy via Unix socket |
| **Redis** | ✅ Private Network | VPC-internal only |
| **API Endpoints** | ✅ JWT Required | All protected routes need Bearer token |
| **Admin Endpoints** | ✅ RBAC | Requires admin role + JWT |
| **Secrets** | ✅ Secret Manager | IAM-based access control |

## 🚀 Implementation Plan

### Week 1: Critical Security Fixes

```bash
# 1. Create admin setup guard
create backend/middleware/admin_setup_guard.go

# 2. Update routes
update backend/cmd/server/main.go

# 3. Protect skill creation
move POST /skills/taxonomy to protected routes

# 4. Deploy
make build-push && make deploy-app
```

### Week 2: CORS & Rate Limiting

```bash
# 1. CORS whitelist
update middleware/cors.go with allowed origins

# 2. Per-user rate limiting
implement Redis-backed rate limits

# 3. Deploy
make build-push && make deploy-app
```

### Week 3: Monitoring & Alerts

```bash
# 1. Set up Cloud Monitoring dashboards
# 2. Configure alerts for:
#    - Failed auth attempts spike
#    - Unusual traffic patterns
#    - Error rate threshold exceeded
# 3. Set up log-based metrics
```

## 🔍 Quick Security Audit Commands

```bash
# Check public endpoints
curl https://civicweave-backend-162941711179.us-central1.run.app/health

# Try protected endpoint without auth (should get 401)
curl https://civicweave-backend-162941711179.us-central1.run.app/api/me

# Check CORS headers
curl -H "Origin: https://evil.com" \
     -H "Access-Control-Request-Method: POST" \
     -X OPTIONS \
     https://civicweave-backend-162941711179.us-central1.run.app/api/auth/login

# Test rate limiting
for i in {1..20}; do 
  curl -X POST https://civicweave-backend-162941711179.us-central1.run.app/api/auth/login \
    -d '{"email":"test@test.com","password":"wrong"}'
done
```

## 📝 Notes

### Why Backend Must Be Public
- Web applications need to call the API from browsers
- Can't use IAM authentication for browser clients
- Must rely on application-level auth (JWT)

### Why This Is Secure Enough
- ✅ JWT authentication on all sensitive endpoints
- ✅ RBAC for admin operations
- ✅ Rate limiting prevents brute force
- ✅ Database not publicly accessible
- ✅ Secrets managed securely

### Additional Hardening (Optional)
- Consider Cloudflare in front of Cloud Run
- Implement request signing
- Add honeypot endpoints to detect scanners
- Geographic restrictions if known user base

