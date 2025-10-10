# Repository Guidelines

## Project Structure & Module Organization
CivicWeave is split into a Go API inside `backend/` and a Vite-powered React client inside `frontend/`. API entrypoints live in `backend/cmd/server`, with supporting packages under `handlers/`, `services/`, and `models/`. Database schema changes belong in `backend/migrations/`, and reusable SQL helpers stay in `database/`. Frontend features sit in `frontend/src/` and group UI, hooks, and tests by feature; shared styles and Tailwind config live under `frontend/src/styles/` and `tailwind.config.js`. Infrastructure-as-code lives in `infrastructure/terraform` and Docker orchestration is handled by the root `docker-compose.yml`.

## Build, Test, and Development Commands
Use `make dev-up` to boot Postgres, Redis, and both apps in Docker; `make dev-down` stops the stack. When iterating only on Go, run `go run cmd/server/main.go` from `backend/` with a local `.env`. Frontend hot reload is available through `npm run dev` (from `frontend/`). Database migrations and sample data are applied via `make db-migrate` and `make db-seed`. CI-equivalent checks are `make lint` and `make test`.

## Coding Style & Naming Conventions
Run `golangci-lint run` and `gofmt` (via `go fmt ./...`) before committing Go changes; keep handlers thin and place business logic in `services`. Go package names stay lowercase and match their directory. The frontend follows ESLint + Prettier defaults (`npm run lint`), React components use PascalCase, hooks use `useThing` camelCase, and Tailwind classes should be composed with `clsx` or `tailwind-merge` helpers for readability.

## Testing Guidelines
Go tests mirror package names (e.g., `handlers/user_test.go`) and should mock external services. Prefer table-driven tests for handlers and services. Frontend unit tests live alongside components or in `frontend/src/__tests__/` using Vitest; name files `ComponentName.test.jsx`. Run `make test` before opening a PR, and ensure new features include at least one automated test touching the happy path.

## Commit & Pull Request Guidelines
Follow the existing `Type: Message` convention (`Fix: Correct frontend API URL configuration`). Keep commits scoped to a single concern and include migration or seed IDs in the message when relevant. PRs should link Jira/GitHub issues, describe the user-facing change, and attach screenshots for UI updates or curl examples for new API endpoints. Request a reviewer from both backend and frontend when changes span layers.

## Environment & Security Tips
Duplicate `env.example` or `backend/env.example` when configuring local secrets; never check real credentials into git. Prefer Docker secrets or Secret Manager for production values, and verify CORS origin lists whenever new environments are introduced.
