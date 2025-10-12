# Deployment Notes

## ⚠️ Important: Why We Changed deploy-app

### The Problem
Previously, `make deploy-app` used Terraform to deploy:
```bash
terraform apply -target=google_cloud_run_service.backend
terraform apply -target=google_cloud_run_service.frontend
```

**This didn't work** because:
- Terraform only checks if the **configuration** changed
- When using `:latest` tag, the image reference in Terraform stays the same
- Even though we pushed a new image, Terraform sees "no changes needed"
- Cloud Run never pulls the new image!

### The Solution
Now `make deploy-app` directly uses `gcloud run deploy`:
```bash
gcloud run deploy civicweave-backend --image=...
gcloud run deploy civicweave-frontend --image=...
```

**This works because:**
- ✅ Forces Cloud Run to create a new revision
- ✅ Always pulls the latest image from the registry
- ✅ Immediately routes traffic to the new revision
- ✅ Works every time, no caching issues

## 🚀 Correct Deployment Workflow

### Complete Deployment (Code Changes)
```bash
# 1. Build and push new images with code changes
make build-push

# 2. Deploy to Cloud Run (forces new revisions)
make deploy-app
```

### Infrastructure Changes Only
```bash
# Use Terraform for infrastructure/configuration changes
make deploy-infra
```

## 📋 What Each Command Does

| Command | Purpose | When to Use |
|---------|---------|-------------|
| `make build-dev` | Build local Docker images | Local testing only |
| `make build-push` | Build & push to GCP registry | Before every production deployment |
| `make deploy-app` | Deploy apps to Cloud Run | After `build-push` |
| `make deploy-infra` | Terraform infrastructure | Database, Redis, secrets, IAM changes |

## ✅ Current Production Status

After the latest deployment:
- **Backend**: `civicweave-backend-00030-pl4` 
  - URL: https://civicweave-backend-162941711179.us-central1.run.app
  
- **Frontend**: `civicweave-frontend-00016-tob`
  - URL: https://civicweave-frontend-162941711179.us-central1.run.app
  - ✅ Now correctly points to production backend
  - ✅ Registration bug fixed (snake_case fields)

## 🔍 How to Verify Deployment

### Check Active Revisions
```bash
gcloud run revisions list --service=civicweave-backend --region=us-central1
gcloud run revisions list --service=civicweave-frontend --region=us-central1
```

### Check Frontend API URL
```bash
# Should show production backend URL, not localhost
curl -s https://civicweave-frontend-162941711179.us-central1.run.app/assets/*.js | \
  grep -o 'civicweave-backend.*run\.app' | head -1
```

### Test Registration
Open: https://civicweave-frontend-162941711179.us-central1.run.app/register

Should:
- ✅ Load without errors
- ✅ Not show localhost:8080 in console
- ✅ Successfully register new users (no 400/500 errors)

## 🐛 Troubleshooting

### Frontend still shows localhost in browser console
**Cause**: Browser cached old JavaScript files

**Solution**:
1. Hard refresh: `Ctrl+Shift+R` (Windows/Linux) or `Cmd+Shift+R` (Mac)
2. Or clear browser cache
3. Check in incognito/private window

### Deployment succeeds but changes not visible
**Cause**: Old revision still serving traffic

**Solution**:
```bash
# Check which revision is active
gcloud run services describe civicweave-frontend --region=us-central1

# If needed, manually migrate traffic
gcloud run services update-traffic civicweave-frontend \
  --to-latest --region=us-central1
```

### Build succeeds but deploy fails
**Cause**: Permission issues or wrong project

**Solution**:
```bash
# Verify you're in the right project
gcloud config get-value project

# Should be: civicweave-474622
```

## 📝 Quick Reference

### One-Line Full Deployment
```bash
make build-push && make deploy-app
```

### Check Deployment Status
```bash
gcloud run services list --region=us-central1
```

### View Recent Logs
```bash
gcloud run services logs read civicweave-frontend --region=us-central1 --limit=50
gcloud run services logs read civicweave-backend --region=us-central1 --limit=50
```

### Rollback to Previous Revision
```bash
# List revisions
gcloud run revisions list --service=civicweave-frontend --region=us-central1

# Rollback to specific revision
gcloud run services update-traffic civicweave-frontend \
  --to-revisions=civicweave-frontend-00015-srb=100 \
  --region=us-central1
```

