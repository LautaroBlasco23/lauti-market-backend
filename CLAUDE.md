# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Development
make dev          # Run with hot reload (air)
make start        # Interactive start (choose: dev, docker test, prod)

# Docker
make docker-up    # Build and start full stack (API + PostgreSQL + Redis)
make docker-down  # Stop full stack
make db-up        # Start database services only (PostgreSQL + Redis)
make db-down      # Stop database services

# Data
make inject-data  # Seed fake stores and products into the running app (go run ./cmd/seed)

# Code quality
make code-check   # Run gofumpt + golangci-lint
make test         # Run tests with gotestsum (short-verbose format)
make test-security # Run security tests against running containers

# Install dev tools
make install-tools  # Install gofumpt, golangci-lint, air, gotestsum
```

Server runs on `:8080` (Docker) or `PORT` env variable (default `:8000` in dev).

## Architecture

Clean Architecture with these layers per module (dependency direction: infrastructure → application → domain):

```
internal/
├── api/                    # Shared infrastructure (errors, CORS, UUID, validator)
└── {module}/               # auth, user, store, product
    ├── domain/             # Entities with validation + repository interfaces
    ├── application/        # Use cases / service layer
    └── infrastructure/
        ├── controller/     # HTTP handlers
        ├── dto/            # Request/response types
        ├── repository/     # PostgreSQL implementations
        ├── routes/         # Route registration
        ├── utils/          # Module-specific utilities
        └── wiring.go       # Dependency injection entry point
```

**Entry point:** `cmd/api/main.go` — initializes DB, sets up CORS middleware, then calls each module's `Wire()` in order: User → Store → Auth → Product.

**Adding a new module:** create `domain/` (entity + repository interface), `application/service.go` (use cases), `infrastructure/` (controller, dto, repository, routes, wiring.go), then call `Wire()` from `main.go`.

## Key Patterns

**Error handling:** Centralized error constants live in `internal/api/domain/errors.go`. Each module imports from `apiDomain` (not its own domain package) for shared errors.

**Validation:** Domain entities carry struct tags validated via `go-playground/validator`. The shared validator wrapper is at `internal/api/infrastructure/validator.go`.

**Dependency injection:** Each module's `wiring.go` instantiates all layers and registers routes. Interfaces are used at every boundary (repository, ID generator, password hasher, token generator).

**Database:** Raw SQL with `database/sql` + `lib/pq`. No ORM. Migrations are a single file: `migrations/init.sql`. Schema: `users`, `stores`, `auths` (account_type CHECK: "user"|"store"), `products`.

**Auth:** JWT-based. `AuthService` depends on both `UserService` and `StoreService` to create the underlying account before creating the auth record.

## Environment Variables

```
DB_HOST, DB_USER, DB_PASSWORD, DB_NAME, DB_PORT, DB_SSLMODE
JWT_SECRET
PORT
UNSPLASH_ACCESS_KEY   # optional — used by cmd/seed to attach product images
```
