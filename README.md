# CivicWeave MVP

A volunteer management platform for connecting volunteers with civic initiatives.

## Architecture

- **Backend**: Golang + Gin framework (REST API)
- **Frontend**: React + Vite + React Router
- **Database**: PostgreSQL + Redis
- **Auth**: JWT with dual support (email/password + Google OAuth)
- **Email**: Mailgun integration
- **Infrastructure**: GCP (Cloud Run, Cloud SQL, Memorystore)

## Quick Start (Local Development)

1. **Prerequisites**:
   - Docker & Docker Compose
   - Go 1.21+
   - Node.js 18+
   - Terraform (for infrastructure)

2. **Setup**:
   ```bash
   # Clone and start services
   git clone <repo-url>
   cd CivicWeave
   
   # Start local development environment
   make dev-up
   
   # Initialize database
   make db-migrate
   
   # Seed admin user
   make db-seed
   ```

3. **Access**:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - Admin Portal: http://localhost:3000/admin

## Development Commands

```bash
make dev-up          # Start all services
make dev-down        # Stop all services
make db-migrate      # Run database migrations
make db-seed         # Seed initial data
make test           # Run tests
make lint           # Run linters
```

## Project Structure

```
/
├── backend/         # Golang API server
├── frontend/        # React SPA
├── infrastructure/  # Terraform configs
├── shared/         # Shared types/schemas
└── docker-compose.yml
```

## Deployment

### GCP Project Setup (One-time)

1. **Set up new GCP project**:
   ```bash
   make setup-gcp
   ```
   This creates a new GCP project, enables APIs, and generates secure configuration.

2. **Update API keys** in `infrastructure/terraform/terraform.tfvars`:
   - Mailgun API key and domain
   - Google OAuth client credentials

3. **Deploy infrastructure**:
   ```bash
   make deploy-infra
   ```

4. **Build and deploy applications**:
   ```bash
   make build-push
   make deploy-app
   ```

See `infrastructure/terraform/README.md` for detailed deployment instructions.

## Contributing

1. Create feature branch from `main`
2. Make changes
3. Test locally with `make test`
4. Submit pull request

## License

MIT
