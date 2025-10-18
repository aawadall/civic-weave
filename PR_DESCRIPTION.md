# Chore: Improve auth logging, externalize CORS config, and document Terraform state backup

## Summary

This PR implements three infrastructure improvements to enhance debugging, security, and deployment practices:

1. **Auth Error Logging Enhancement**: Add structured error logging to Register handler for better 502 debugging
2. **CORS Configuration Externalization**: Move CORS origins to environment variables with proper header handling  
3. **Terraform State Management**: Document backup process and add remote backend configuration

## 🤖 GitHub Copilot Review

**Status**: ✅ **Approved with 2 improvements implemented**

Copilot reviewed all 11 changed files and provided valuable feedback:
- **Critical Fix**: Fixed JSON structure mismatch in Terraform state recovery commands
- **Code Quality**: Extracted long CORS default string to constant for better maintainability
- **Overall Assessment**: Well-structured changes with comprehensive documentation

## Changes Made

### 1. Auth Error Logging (`backend/handlers/auth_full.go`)
- Added structured logging for database errors in user registration
- Lines 85-88: Log `REGISTER_DB_ERROR` when checking existing users
- Lines 109-111: Log `REGISTER_USER_CREATE_ERROR` when creating users
- Maintains existing error responses while adding debugging information

### 2. CORS Configuration Externalization
- **`backend/config/config.go`**: Added `CORSConfig` struct and parsing logic
- **`backend/middleware/cors.go`**: Updated to accept origins parameter and set headers on all response paths
- **`backend/cmd/server/main.go`**: Wire CORS config through to middleware
- **`backend/env.example`**: Added `CORS_ALLOWED_ORIGINS` environment variable

### 3. Terraform State Management
- **`infrastructure/terraform/STATE_BACKUP.md`**: Created comprehensive backup documentation
- **`infrastructure/terraform/README.md`**: Added state management section with remote backend guidance
- **`infrastructure/terraform/main.tf`**: Added commented remote backend configuration
- Terraform state files already properly ignored by git

### 4. Copilot-Recommended Improvements
- **`infrastructure/terraform/STATE_BACKUP.md`**: Fixed JSON structure mismatch in recovery commands (removed `.value` suffix)
- **`backend/config/config.go`**: Extracted long CORS default string to `defaultCORSOrigins` constant for better maintainability

## Testing Instructions

### 1. Auth Error Logging
```bash
# Start the backend in development mode
cd backend && go run cmd/server/main.go

# Attempt registration with invalid data to trigger database errors
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"short","name":"Test","consent_given":true}'

# Check logs for structured error messages:
# ❌ REGISTER_DB_ERROR: Failed to check existing user for email test@example.com: [error details]
```

### 2. CORS Configuration
```bash
# Test with different origins
curl -H "Origin: http://localhost:3000" http://localhost:8080/api/health
curl -H "Origin: https://civicweave.com" http://localhost:8080/api/health
curl -H "Origin: https://malicious-site.com" http://localhost:8080/api/health

# Verify CORS headers are present in all responses (including errors)
```

### 3. Environment Variable Testing
```bash
# Test with custom CORS origins
export CORS_ALLOWED_ORIGINS="http://localhost:3000,https://my-custom-domain.com"
cd backend && go run cmd/server/main.go

# Verify only specified origins are allowed
```

## Deployment Instructions

### 1. Terraform State Backup (Before Deployment)
```bash
# Backup current Terraform state to GCloud Secret Manager
cd infrastructure/terraform
gcloud secrets create terraform-outputs-backup --data-file=<(terraform output -json) --project=civicweave-474622
gcloud secrets create terraform-state-full-backup --data-file=terraform.tfstate --project=civicweave-474622
```

### 2. Deploy Infrastructure (if needed)
```bash
# Deploy infrastructure changes
make deploy-infra
```

### 3. Configure Cloud Run with CORS Environment Variable
```bash
# Add CORS_ALLOWED_ORIGINS to Cloud Run configuration
gcloud run services update civicweave-backend \
  --region=us-central1 \
  --set-env-vars="CORS_ALLOWED_ORIGINS=https://civicweave.com,https://civicweave-frontend-162941711179.us-central1.run.app" \
  --quiet
```

### 4. Build and Deploy Applications
```bash
# Build and push container images
make build-push

# Deploy to Cloud Run
make deploy-app
```

### 5. Verify Deployment
```bash
# Check backend health
curl https://civicweave-backend-162941711179.us-central1.run.app/health

# Test CORS headers
curl -H "Origin: https://civicweave.com" https://civicweave-backend-162941711179.us-central1.run.app/health
```

## Configuration Migration Guide

### For Existing Deployments
1. **Backup Terraform state**: `cd infrastructure/terraform && gcloud secrets create terraform-outputs-backup --data-file=<(terraform output -json) --project=civicweave-474622`
2. **Add CORS environment variable** to Cloud Run: Use the gcloud command in deployment instructions
3. **Deploy applications**: `make build-push && make deploy-app`
4. **Monitor logs** for improved error visibility

### For New Deployments
1. **Set up GCP project**: `make setup-gcp`
2. **Deploy infrastructure**: `make deploy-infra`
3. **Configure Cloud Run**: `make configure-cloud-run`
4. **Set CORS_ALLOWED_ORIGINS** in Cloud Run environment
5. **Deploy applications**: `make build-push && make deploy-app`
6. **Use remote backend** for Terraform state (uncomment backend config in main.tf)
7. **Follow STATE_BACKUP.md** for state management best practices

## Security Considerations

- **CORS origins are now configurable** - ensure production origins are properly set
- **Terraform state backup** - critical infrastructure data is preserved in GCloud Secret Manager
- **Error logging** - sensitive information is not logged, only error types and context

## Rollback Plan

If issues arise:
1. **Revert CORS changes**: Remove `CORS_ALLOWED_ORIGINS` env var to use defaults
2. **Auth logging**: No impact on functionality, only adds logging
3. **Terraform state**: Use backup from GCloud Secret Manager if needed

## Monitoring

After deployment, monitor:
- **Backend logs** for new structured error messages
- **CORS headers** in browser network tab
- **Terraform state** backup completion in GCloud Secret Manager