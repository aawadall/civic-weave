.PHONY: dev-up dev-down db-migrate db-seed test lint clean

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
	if [ -f .env ]; then set -a && . .env && set +a; fi && \
	terraform init && terraform apply

deploy-app:
	cd infrastructure/terraform && \
	if [ -f .env ]; then set -a && . .env && set +a; fi && \
	terraform apply -target=google_cloud_run_service.backend && \
	terraform apply -target=google_cloud_run_service.frontend

build-push:
	cd backend && docker build -t us-central1-docker.pkg.dev/civicweave-474622/civicweave/backend:latest .
	cd backend && docker push us-central1-docker.pkg.dev/civicweave-474622/civicweave/backend:latest
	cd frontend && docker build -t us-central1-docker.pkg.dev/civicweave-474622/civicweave/frontend:latest .
	cd frontend && docker push us-central1-docker.pkg.dev/civicweave-474622/civicweave/frontend:latest

# Help
help:
	@echo "Available commands:"
	@echo "  dev-up        - Start development environment"
	@echo "  dev-down      - Stop development environment"
	@echo "  dev-logs      - View logs"
	@echo "  db-migrate    - Run database migrations"
	@echo "  db-seed       - Seed initial data"
	@echo "  db-backfill-skills - Convert existing JSONB skills to vector claims"
	@echo "  db-reset      - Reset database with fresh data"
	@echo "  test          - Run tests"
	@echo "  lint          - Run linters"
	@echo "  clean         - Clean up containers and volumes"
	@echo "  setup-gcp     - Set up new GCP project (run once)"
	@echo "  deploy-infra  - Deploy infrastructure to GCP"
	@echo "  deploy-app    - Deploy applications to GCP"
	@echo "  build-push    - Build and push container images"
