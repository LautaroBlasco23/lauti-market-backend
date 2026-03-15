# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Install dev tools (gofumpt, golangci-lint, air, gotestsum)
make install-tools

# Start databases (PostgreSQL + Redis via Docker)
make db-up

# Start development server with hot reload (air) - requires db-up first
make dev

# Format and lint
make code-check

# Run tests
make test

# Full stack in Docker
make docker-up && make docker-down
```

The API server runs on `:8080`.

## Architecture

Each module (`auth`, `user`, `store`, `product`, `order`) follows the same three-layer structure:

```
domain/         → entities, repository interfaces, domain errors
application/    → service (business logic, orchestration)
infrastructure/ → controller (HTTP handlers + DTOs), repository (PostgreSQL impl), routes/, wiring.go
```

Dependency rule: Infrastructure → Application → Domain. Domain has no external dependencies.

Shared cross-module code lives in `internal/api/`:
- `domain/errors.go` — centralized domain errors used across all modules
- `domain/id_generator.go` — IDGenerator interface
- `infrastructure/` — CORS middleware, UUID generator, request validator, JWT auth middleware

### Module wiring

Each module exposes a `Wire(...)` function that constructs the full dependency chain (repo → service → controller → routes) and registers HTTP routes on the provided `*http.ServeMux`. Main wiring happens in `cmd/api/main.go`. The `order` module's `Wire` takes additional args: `productRepo` and `authMw` (JWT middleware).

### HTTP layer

Uses Go's standard `http.ServeMux` with Go 1.22+ route patterns (e.g., `"POST /auth/register/user"`). No external router. Controllers map domain errors to HTTP status codes.

### Database

PostgreSQL only (Redis is configured but unused). Schema lives in `migrations/init.sql`. The `database/` package exposes a `Database` interface with a `postgres.go` implementation.

## Key dependencies

- `github.com/golang-jwt/jwt/v5` — JWT auth
- `github.com/google/uuid` — ID generation
- `golang.org/x/crypto` — bcrypt hashing
- `github.com/go-playground/validator` — request validation
- `github.com/lib/pq` — PostgreSQL driver
