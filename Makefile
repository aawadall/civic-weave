.PHONY: dev-up dev-down db-migrate db-seed test lint clean fetch-secrets setup-gcp deploy-infra deploy-app configure-cloud-run build-dev build-push build-push-prod help

# Development commands
dev-up:
	docker-compose up -d postgres redis
	sleep 5
	docker-compose up -d backend frontend

dev-down:
	docker-compose down

dev-logs:
	docker-compose logs -f

# Database commands
db-migrate:
	cd backend && go run cmd/migrate/main.go up

db-seed:
	cd backend && go run cmd/seed/main.go

db-backfill-skills:
	cd backend && go run scripts/backfill_skill_vectors.go

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
	cd backend && docker build -t civicweave_backend:dev .
	cd frontend && docker build \
		--build-arg VITE_API_BASE_URL=http://localhost:8081/api \
		--build-arg VITE_GOOGLE_CLIENT_ID=$${GOOGLE_CLIENT_ID:-162941711179-5ducggubvulr92290a5qasgupdr7ifqk.apps.googleusercontent.com} \
		-t civicweave_frontend:dev .

# Build and push for production
build-push:
	cd backend && docker build -t us-central1-docker.pkg.dev/civicweave-474622/civicweave/backend:latest .
	cd backend && docker push us-central1-docker.pkg.dev/civicweave-474622/civicweave/backend:latest
	cd frontend && docker build \
		--build-arg VITE_API_BASE_URL=https://civicweave-backend-162941711179.us-central1.run.app/api \
		--build-arg VITE_GOOGLE_CLIENT_ID=$${GOOGLE_CLIENT_ID:-162941711179-5ducggubvulr92290a5qasgupdr7ifqk.apps.googleusercontent.com} \
		-t us-central1-docker.pkg.dev/civicweave-474622/civicweave/frontend:latest .
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
	@echo "  dev-logs      - View logs"
	@echo "  build-dev     - Build Docker images for local testing"
	@echo ""
	@echo "Database:"
	@echo "  db-migrate    - Run database migrations"
	@echo "  db-seed       - Seed initial data"
	@echo "  db-backfill-skills - Convert existing JSONB skills to vector claims"
	@echo "  db-reset      - Reset database with fresh data"
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
