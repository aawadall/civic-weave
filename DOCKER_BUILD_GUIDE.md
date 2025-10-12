# Docker Build Guide

## Overview

We have two separate build configurations for CivicWeave:

1. **Development builds** - For local testing with Docker
2. **Production builds** - For deploying to Google Cloud Run

## 🔧 Development Build

### When to use:
- Testing locally with Docker containers
- Want to use Docker but point to local backend

### Command:
```bash
make build-dev
```

### What it does:
- Builds backend: `civicweave_backend:dev`
- Builds frontend: `civicweave_frontend:dev`
- Frontend configured to call: `http://localhost:8081`
- **Does NOT push to Google Container Registry**

### Usage with docker-compose:
```bash
# docker-compose.yml is already configured for dev
make dev-up
```

The `docker-compose.yml` file automatically builds with localhost URLs.

---

## 🚀 Production Build

### When to use:
- Deploying to production (Google Cloud Run)
- After making code changes that need to go live

### Command:
```bash
make build-push
# or
make build-push-prod  # alias
```

### What it does:
- Builds backend: `us-central1-docker.pkg.dev/civicweave-474622/civicweave/backend:latest`
- Builds frontend: `us-central1-docker.pkg.dev/civicweave-474622/civicweave/frontend:latest`
- Frontend configured to call: `https://civicweave-backend-162941711179.us-central1.run.app`
- **Pushes to Google Artifact Registry**

### Complete production deployment:
```bash
# 1. Build and push new images
make build-push

# 2. Deploy to Cloud Run
make deploy-app

# OR use gcloud directly to force new revision
gcloud run deploy civicweave-backend --image=us-central1-docker.pkg.dev/civicweave-474622/civicweave/backend:latest --region=us-central1 --platform=managed
gcloud run deploy civicweave-frontend --image=us-central1-docker.pkg.dev/civicweave-474622/civicweave/frontend:latest --region=us-central1 --platform=managed
```

---

## 📋 Comparison Table

| Feature | Development | Production |
|---------|------------|------------|
| **Backend Image** | `civicweave_backend:dev` | `us-central1-docker.pkg.dev/.../backend:latest` |
| **Frontend Image** | `civicweave_frontend:dev` | `us-central1-docker.pkg.dev/.../frontend:latest` |
| **API URL** | `http://localhost:8081` | `https://civicweave-backend-*.run.app` |
| **Push to Registry** | ❌ No | ✅ Yes |
| **Use Case** | Local Docker testing | Production deployment |
| **Command** | `make build-dev` | `make build-push` |

---

## 🎯 Quick Workflows

### Local Development (Native - No Docker)
```bash
# Terminal 1: Start backend
cd backend
go run cmd/server/main.go

# Terminal 2: Start frontend
cd frontend
npm run dev

# Access: http://localhost:3000
```

### Local Development (Docker)
```bash
# Build dev images (optional, docker-compose builds automatically)
make build-dev

# Start everything
make dev-up

# Access: http://localhost:3001
```

### Production Deployment
```bash
# Full deployment
make build-push
make deploy-app

# OR just redeploy without rebuilding
make deploy-app
```

---

## 🔑 Environment Variables

### Development
- Hardcoded in `docker-compose.yml`
- API points to `localhost:8081`

### Production
- Set via `infrastructure/terraform/.env`
- Managed by Google Secret Manager
- API points to production Cloud Run URL

---

## ⚠️ Important Notes

1. **Vite bakes in environment variables at build time**
   - Changing `.env` won't affect already-built images
   - Must rebuild to pick up new API URLs

2. **docker-compose.yml is for development only**
   - Already configured with localhost URLs
   - Rebuilds happen automatically on `docker-compose up --build`

3. **Production images require correct build args**
   - `VITE_API_BASE_URL` must point to production backend
   - Set in Makefile's `build-push` target

4. **After code changes**
   - Development: Just restart containers or rebuild
   - Production: Must run `make build-push` then deploy

---

## 🆘 Troubleshooting

### Frontend shows 404 errors to localhost in production
**Problem**: Frontend built with wrong API URL

**Solution**:
```bash
make build-push  # Rebuild with production URL
make deploy-app  # Redeploy
```

### Changes not appearing in production
**Problem**: Old Docker image cached or not pushed

**Solution**:
```bash
make build-push  # Force rebuild and push
gcloud run deploy civicweave-frontend --image=...  # Force new revision
```

### Local frontend can't reach backend
**Problem**: Port mismatch or backend not running

**Solution**:
```bash
# Check ports
docker-compose ps

# Backend should be on 8081
# Frontend should be on 3001
```

