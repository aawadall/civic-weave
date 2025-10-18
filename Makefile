.PHONY: dev-up dev-down db-migrate db-seed test lint clean fetch-secrets setup-gcp deploy-infra deploy-app configure-cloud-run build-dev build-push build-push-prod help db-agent-dev db-client-ping db-deploy-v3 db-keygen

# Development commands
dev-up:
	docker-compose up -d postgres redis
	sleep 5
	docker-compose up -d backend frontend

dev-down:
	docker-compose down

dev-rebuild:
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d postgres redis
	sleep 5
	docker-compose up -d backend frontend

dev-logs:
	docker-compose logs -f

# Database commands
db-migrate:
	cd backend && go run cmd/migrate/main.go up

db-seed:
	cd backend && go run cmd/seed/main.go

db-backfill-skills:
	cd backend && go run scripts/backfill_skill_vectors.go

# Enhanced migration system (v2)
db-migrate-v2:
	cd backend && go run cmd/migrate/main.go -command=up -runtime-version=1.0.0

db-migrate-v2-prod:
	cd backend && go run cmd/migrate/main.go -command=up -runtime-version=1.0.0 -env=../.env.production

db-migrate-status:
	cd backend && go run cmd/migrate/main.go -command=status

db-migrate-status-prod:
	cd backend && go run cmd/migrate/main.go -command=status -env=../.env.production

db-migrate-compat:
	cd backend && go run cmd/migrate/main.go -command=compatibility -runtime-version=1.0.0

db-migrate-validate:
	cd backend && go run cmd/migrate/main.go -command=validate

db-migrate-check:
	cd backend && go run cmd/migrate/main.go -command=check -runtime-version=1.0.0

db-migrate-rollback:
	@read -p "Enter target version (e.g., 1.0.0): " version; \
	cd backend && go run cmd/migrate/main.go -command=down -version=$$version

# Schema state validation
db-schema-state:
	cd backend && go run cmd/migrate/main.go -command=schema-state

db-drift-detect:
	cd backend && go run cmd/migrate/main.go -command=drift-detect

db-validate-state:
	@read -p "Enter target version (e.g., 1.0.0): " version; \
	cd backend && go run cmd/migrate/main.go -command=validate-state -version=$$version

# Remote database deployment
db-deploy-status:
	./scripts/deploy-db.sh --status

db-deploy-dry:
	./scripts/deploy-db.sh --dry-run

db-deploy:
	./scripts/deploy-db.sh

db-deploy-version:
	@read -p "Enter target version (e.g., 011): " version; \
	./scripts/deploy-db.sh --version $$version

db-rollback:
	@read -p "Enter rollback version (e.g., 010): " version; \
	./scripts/deploy-db.sh --rollback $$version

# gRPC Database Agent System (v3)
db-agent-dev:
	cd backend && go run cmd/db-agent/main.go -port=50051 -host=localhost

db-client-ping:
	cd backend && go run cmd/db-client/main.go -command=ping -agent=localhost:50051

db-deploy-v3:
	cd backend && go run cmd/db-client/main.go -command=deploy -manifest=./manifest -database=prod -agent=localhost:50051

db-keygen:
	cd backend && go run cmd/db-keygen/main.go -server -description="Development Server Key"

db-keygen-client:
	cd backend && go run cmd/db-keygen/main.go -client -agent-url=localhost:50051 -description="Development Client Key"

db-compare-v3:
	cd backend && go run cmd/db-client/main.go -command=compare -manifest=./manifest -database=prod -agent=localhost:50051

db-download-v3:
	cd backend && go run cmd/db-client/main.go -command=download -database=prod -output=./manifest -agent=localhost:50051

db-history-v3:
	cd backend && go run cmd/db-client/main.go -command=history -database=prod -agent=localhost:50051

db-bootstrap-v3:
	cd backend && go run cmd/db-client/main.go -command=bootstrap -manifest=./manifest -database=new-db -agent=localhost:50051

# Cloud Deployment Commands
db-agent-docker:
	cd backend && ./scripts/setup-agent.sh setup

db-agent-docker-stop:
	cd backend && ./scripts/setup-agent.sh stop

db-agent-docker-logs:
	cd backend && ./scripts/setup-agent.sh logs

db-agent-k8s:
	cd backend && ./scripts/deploy-k8s.sh deploy

db-agent-k8s-update:
	cd backend && ./scripts/deploy-k8s.sh update

db-agent-k8s-status:
	cd backend && ./scripts/deploy-k8s.sh status

db-agent-k8s-cleanup:
	cd backend && ./scripts/deploy-k8s.sh cleanup

# Google Cloud Run Deployment
db-agent-gcloud:
	cd backend && ./scripts/deploy-gcloud.sh deploy

db-agent-gcloud-update:
	cd backend && ./scripts/deploy-gcloud.sh update

db-agent-gcloud-status:
	cd backend && ./scripts/deploy-gcloud.sh status

db-agent-gcloud-logs:
	cd backend && ./scripts/deploy-gcloud.sh logs

db-agent-gcloud-cleanup:
	cd backend && ./scripts/deploy-gcloud.sh cleanup

# Batch jobs
job-calculate-matches:
	cd backend && venv/bin/python jobs/calculate_matches.py

job-notify-matches:
	cd backend && ./jobs/run_daily_matching.sh

job-setup-python:
	cd backend && python3 -m venv venv && venv/bin/pip install psycopg2-binary python-dotenv numpy

matching-worker:
	cd backend && go run cmd/matchingworker/main.go

db-reset:
	docker-compose down -v
	docker-compose up -d postgres
	sleep 5
	make db-migrate
	make db-seed

# Testing
test:
	cd backend && go test ./...
	cd frontend && npm test

lint:
	cd backend && golangci-lint run
	cd frontend && npm run lint

# Cleanup
clean:
	docker-compose down -v
	docker system prune -f

clean-frontend:
	docker-compose down frontend
	docker rmi civicweave_frontend:dev 2>/dev/null || true
	docker-compose build --no-cache frontend
	docker-compose up -d frontend

# GCP Project Setup
setup-gcp:
	cd infrastructure/terraform && ./setup.sh

# Production deployment
deploy-infra:
	cd infrastructure/terraform && \
	if [ -f .env ]; then export $$(cat .env | grep -v '^#' | xargs); fi && \
	terraform init && terraform apply

configure-cloud-run:
	@echo "⚙️  Configuring Cloud Run environment variables..."
	@echo ""
	@echo "🔍 Discovering Cloud SQL instance..."
	$(eval CLOUD_SQL_CONNECTION := $(shell gcloud sql instances list --format="value(connectionName)" --filter="name:civicweave-postgres" --limit=1))
	@if [ -z "$(CLOUD_SQL_CONNECTION)" ]; then \
		echo "❌ Error: Could not find Cloud SQL instance 'civicweave-postgres'"; \
		exit 1; \
	fi
	@echo "✅ Found: $(CLOUD_SQL_CONNECTION)"
	@echo ""
	@echo "🔐 Configuring backend with secrets and environment variables..."
	gcloud run services update civicweave-backend \
		--region=us-central1 \
		--add-cloudsql-instances=$(CLOUD_SQL_CONNECTION) \
		--set-secrets="JWT_SECRET=jwt-secret:latest,DB_PASSWORD=db-password:latest,MAILGUN_API_KEY=mailgun-api-key:latest,MAILGUN_DOMAIN=mailgun-domain:latest,GOOGLE_CLIENT_ID=google-client-id:latest,GOOGLE_CLIENT_SECRET=google-client-secret:latest,OPENAI_API_KEY=openai-api-key:latest" \
		--set-env-vars="DB_HOST=/cloudsql/$(CLOUD_SQL_CONNECTION),DB_PORT=5432,DB_NAME=civicweave,DB_USER=civicweave,DB_SSLMODE=disable,ENABLE_EMAIL=false,NOMINATIM_BASE_URL=https://nominatim.openstreetmap.org,OPENAI_EMBEDDING_MODEL=text-embedding-3-small" \
		--quiet
	@echo ""
	@echo "✅ Cloud Run configuration complete!"

deploy-app:
	@echo "🚀 Deploying to Cloud Run..."
	@echo "Backend: civicweave-backend"
	gcloud run deploy civicweave-backend \
		--image=us-central1-docker.pkg.dev/civicweave-474622/civicweave/backend:latest \
		--region=us-central1 \
		--platform=managed \
		--quiet
	@echo ""
	@echo "Frontend: civicweave-frontend"
	gcloud run deploy civicweave-frontend \
		--image=us-central1-docker.pkg.dev/civicweave-474622/civicweave/frontend:latest \
		--region=us-central1 \
		--platform=managed \
		--quiet
	@echo ""
	@echo "✅ Deployment complete!"
	@echo "Backend URL:  https://civicweave-backend-162941711179.us-central1.run.app"
	@echo "Frontend URL: https://civicweave-frontend-162941711179.us-central1.run.app"

# Build for local development (localhost URLs)
build-dev:
	@echo "Incrementing versions..."
	@./scripts/increment-version.sh both
	@echo "Building with versions:"
	@echo "Backend: $$(cat backend/VERSION)"
	@echo "Frontend: $$(cat frontend/VERSION)"
	cd backend && docker build --no-cache \
		--build-arg VERSION=$$(cat VERSION) \
		--build-arg GIT_COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
		--build-arg BUILD_ENV=development \
		-t civicweave_backend:dev .
	cd frontend && docker build --no-cache \
		--build-arg VITE_API_BASE_URL=http://localhost:8081/api \
		--build-arg VITE_GOOGLE_CLIENT_ID=$${GOOGLE_CLIENT_ID:-162941711179-5ducggubvulr92290a5qasgupdr7ifqk.apps.googleusercontent.com} \
		--build-arg VITE_VERSION=$$(cat VERSION) \
		--build-arg VITE_GIT_COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
		--build-arg VITE_BUILD_ENV=development \
		-t civicweave_frontend:dev .

# Build and push for production
build-push:
	@echo "Incrementing versions..."
	@./scripts/increment-version.sh both
	@echo "Building with versions:"
	@echo "Backend: $$(cat backend/VERSION)"
	@echo "Frontend: $$(cat frontend/VERSION)"
	@cp backend/VERSION backend/VERSION.tmp
	cd backend && docker build --no-cache \
		--build-arg VERSION=$$(cat VERSION.tmp) \
		--build-arg GIT_COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
		--build-arg BUILD_ENV=production \
		-t us-central1-docker.pkg.dev/civicweave-474622/civicweave/backend:latest .
	@rm -f backend/VERSION.tmp
	cd backend && docker push us-central1-docker.pkg.dev/civicweave-474622/civicweave/backend:latest
	@cp frontend/VERSION frontend/VERSION.tmp
	cd frontend && docker build --no-cache \
		--build-arg VITE_API_BASE_URL=https://civicweave-backend-162941711179.us-central1.run.app/api \
		--build-arg VITE_GOOGLE_CLIENT_ID=$${GOOGLE_CLIENT_ID:-162941711179-5ducggubvulr92290a5qasgupdr7ifqk.apps.googleusercontent.com} \
		--build-arg VITE_VERSION=$$(cat VERSION.tmp) \
		--build-arg VITE_GIT_COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
		--build-arg VITE_BUILD_ENV=production \
		-t us-central1-docker.pkg.dev/civicweave-474622/civicweave/frontend:latest .
	@rm -f frontend/VERSION.tmp
	cd frontend && docker push us-central1-docker.pkg.dev/civicweave-474622/civicweave/frontend:latest

# Alias for clarity
build-push-prod: build-push

# Secrets management
fetch-secrets:
	./scripts/fetch-secrets.sh

# Help
help:
	@echo "Available commands:"
	@echo ""
	@echo "Development:"
	@echo "  dev-up        - Start development environment (Docker)"
	@echo "  dev-down      - Stop development environment"
	@echo "  dev-rebuild   - Rebuild and restart with no cache"
	@echo "  dev-logs      - View logs"
	@echo "  build-dev     - Build Docker images for local testing"
	@echo "  clean-frontend - Force rebuild frontend container"
	@echo ""
	@echo "Database:"
	@echo "  db-migrate         - Run database migrations"
	@echo "  db-seed            - Seed initial data"
	@echo "  db-backfill-skills - Convert existing JSONB skills to vector claims"
	@echo "  db-reset           - Reset database with fresh data"
	@echo ""
	@echo "Enhanced Migrations (v2):"
	@echo "  db-migrate-v2      - Run enhanced migrations with versioning"
	@echo "  db-migrate-v2-prod - Run enhanced migrations against production database"
	@echo "  db-migrate-status  - Show migration status and pending migrations"
	@echo "  db-migrate-status-prod - Show production migration status"
	@echo "  db-migrate-compat  - Display compatibility matrix"
	@echo "  db-migrate-validate - Validate migration files and integrity"
	@echo "  db-migrate-check   - Check migration health (CI/CD friendly)"
	@echo "  db-migrate-rollback - Rollback to specific version"
	@echo ""
	@echo "Schema State Validation:"
	@echo "  db-schema-state    - Show current database schema state"
	@echo "  db-drift-detect    - Detect schema drift from expected state"
	@echo "  db-validate-state  - Validate database matches intended state for version"
	@echo ""
	@echo "gRPC Database Agent System (v3):"
	@echo "  db-agent-dev      - Run agent server locally"
	@echo "  db-client-ping    - Test connection to agent"
	@echo "  db-keygen         - Generate server API keys"
	@echo "  db-keygen-client  - Generate client API keys"
	@echo "  db-deploy-v3      - Deploy using new gRPC system"
	@echo "  db-compare-v3     - Compare manifest to live database"
	@echo "  db-download-v3    - Download current schema as manifest"
	@echo "  db-history-v3     - Get deployment history"
	@echo "  db-bootstrap-v3   - Initialize new database"
	@echo ""
	@echo "Cloud Deployment:"
	@echo "  db-agent-docker   - Deploy agent with Docker Compose"
	@echo "  db-agent-docker-stop - Stop Docker Compose deployment"
	@echo "  db-agent-docker-logs - View Docker Compose logs"
	@echo "  db-agent-k8s      - Deploy agent to Kubernetes"
	@echo "  db-agent-k8s-update - Update Kubernetes deployment"
	@echo "  db-agent-k8s-status - Show Kubernetes deployment status"
	@echo "  db-agent-k8s-cleanup - Remove Kubernetes deployment"
	@echo ""
	@echo "Google Cloud Run Deployment:"
	@echo "  db-agent-gcloud        - Deploy agent to Cloud Run"
	@echo "  db-agent-gcloud-update - Update Cloud Run deployment"
	@echo "  db-agent-gcloud-status - Show Cloud Run deployment status"
	@echo "  db-agent-gcloud-logs   - View Cloud Run logs"
	@echo "  db-agent-gcloud-cleanup - Remove Cloud Run deployment"
	@echo ""
	@echo "Batch Jobs:"
	@echo "  job-setup-python      - Set up Python environment for batch jobs"
	@echo "  job-calculate-matches - Calculate volunteer-project match scores"
	@echo "  job-notify-matches    - Notify top candidates about project matches"
	@echo ""
	@echo "Testing:"
	@echo "  test          - Run tests"
	@echo "  lint          - Run linters"
	@echo ""
	@echo "Production Deployment:"
	@echo "  fetch-secrets      - Fetch secrets from GCP Secret Manager"
	@echo "  setup-gcp          - Set up new GCP project (run once)"
	@echo "  deploy-infra       - Deploy infrastructure to GCP"
	@echo "  configure-cloud-run - Configure Cloud Run env vars (run once after setup)"
	@echo "  build-push         - Build and push container images to production"
	@echo "  deploy-app         - Deploy applications to GCP Cloud Run"
	@echo ""
	@echo "Cleanup:"
	@echo "  clean         - Clean up containers and volumes"
