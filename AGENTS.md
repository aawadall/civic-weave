# Repository Guidelines

## Project Structure & Module Organization
- Go API lives in `backend/`; the server boots from `backend/cmd/server`, business logic stays in `services/`, HTTP wiring in `handlers/`, and structs in `models/`.
- Database work: migrations reside in `backend/migrations/`, reusable SQL sits in `backend/database/`, and seeds run through `make db-seed`.
- React client sits in `frontend/`; feature folders in `frontend/src/` co-locate components, hooks, and tests, while shared styles land in `frontend/src/styles/` and `tailwind.config.js`.
- Infrastructure code is under `infrastructure/terraform`, with Docker orchestration in the root `docker-compose.yml`.

## Build, Test, and Development Commands
- `make dev-up` / `make dev-down`: boot or stop Postgres, Redis, API, and client containers for full-stack work.
- `go run cmd/server/main.go` (run inside `backend/`): launch the API locally against `.env` without Docker.
- `npm run dev` (run inside `frontend/`): start the Vite dev server.
- `make db-migrate` / `make db-seed`: apply schema changes and load sample fixtures.
- `make lint` / `make test`: execute repo-wide linting and test suites.

## Coding Style & Naming Conventions
- Format Go with `go fmt ./...` and lint via `golangci-lint run`; keep handlers thin and push business rules into `services` packages.
- Go packages and directories stay lowercase and descriptive (e.g., `services/user`).
- React components use PascalCase, hooks follow `useThing`, utilities stay camelCase, and compose Tailwind classes with `clsx` or `tailwind-merge`.

## Testing Guidelines
- Go tests mirror package names (`handlers/user_test.go`), rely on table-driven cases, and mock externals; run with `go test ./...` or `make test`.
- Frontend tests use Vitest beside components or under `frontend/src/__tests__/`; name files `ComponentName.test.jsx` and execute via `npm test` or `make test`.
- Add a happy-path test for every new feature and backfill regressions when bugs surface.

## Commit & Pull Request Guidelines
- Use `Type: Message` commit titles (`Fix: Correct frontend API URL configuration`) and keep each commit focused on one concern; reference migration or seed IDs when relevant.
- Pull requests must link issues, summarize user impact, and attach screenshots for UI or curl examples for API updates.
- Request reviewers from both backend and frontend when changes cross layers, and confirm `make lint` and `make test` succeed before requesting review.

## Security & Configuration Tips
- Duplicate `env.example` or `backend/env.example` for local secrets; never commit real credentials.
- Prefer Docker secrets or a managed secret store in production, and update CORS allowlists when introducing new environments.
