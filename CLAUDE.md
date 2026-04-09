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

# Code quality
make code-check   # Run gofumpt + golangci-lint
make test         # Run all tests with gotestsum (short-verbose format)
make test-coverage # Generate coverage.out and open HTML coverage report
make test-security # Start full Docker stack + run ASVS security tests (cmd/securitytest)

# Install dev tools
make install-tools  # Install gofumpt, golangci-lint, air, gotestsum

# Scripts
make download-images  # Download random product images from Unsplash (requires UNSPLASH_ACCESS_KEY)
```

### Running tests selectively

```bash
# Unit tests only (no external dependencies)
go test ./...

# Include integration and E2E tests (requires Docker for testcontainers)
go test -tags=integration ./...

# Run a specific package
go test ./internal/order/...
```

Server runs on `:8080` (Docker) or `PORT` env variable (default `:8000` in dev).

## Architecture

Clean Architecture with these layers per module (dependency direction: infrastructure → application → domain):

```
internal/
├── api/                    # Shared infrastructure (errors, CORS, UUID, validator)
└── {module}/               # auth, user, store, product, order, image
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

**Special modules:**
- `image` — gRPC client adapter (no HTTP routes), consumed by `product` module
- `order` — `Wire` takes additional args: `productRepo` and `authMw` (JWT middleware)
- `user`, `store`, `product` — `Wire` also takes `authMw` for protecting mutation routes

**Entry point:** `cmd/api/main.go` — initializes DB, sets up CORS middleware, then wires modules in order: User → Store → Auth → Image → Product → Order.

**Adding a new module:** create `domain/` (entity + repository interface), `application/service.go` (use cases), `infrastructure/` (controller, dto, repository, routes, wiring.go), then call `Wire()` from `main.go`.

## Key Patterns

**Error handling:** Centralized error constants live in `internal/api/domain/errors.go`. Each module imports from `apiDomain` (not its own domain package) for shared errors.

**Validation:** Domain entities carry struct tags validated via `go-playground/validator`. The shared validator wrapper is at `internal/api/infrastructure/validator.go`.

**Dependency injection:** Each module's `wiring.go` instantiates all layers and registers routes. Interfaces are used at every boundary (repository, ID generator, password hasher, token generator).

**Database:** Raw SQL with `database/sql` + `lib/pq`. No ORM. Migrations live in `migrations/`. Schema: `users`, `stores`, `auths` (account_type CHECK: "user"|"store"), `products` (with category column), `orders`, `order_items`.

**Auth:** JWT-based. `AuthService` depends on both `UserService` and `StoreService` to create the underlying account before creating the auth record.

**Authentication & ownership enforcement:** `authMw` (JWT middleware) is wired into user, store, and product modules. Pattern:
- Middleware validates JWT and injects account ID into request context
- Controllers check account ownership before allowing PUT/DELETE (user and store modules)
- Controllers check account-type (`user` vs `store`) for store-only operations — only stores can create/update/delete products
- Order read endpoints reject requests from accounts that don't own the order
- Permission denied returns 403; unauthenticated returns 401

**Testing strategy:** Tests are split into four phases:
1. **Unit tests** (phases 1–2) — auth utilities, domain entities, services, JWT/bcrypt, middleware. No external deps.
2. **Shared infra unit tests** (phase 3) — UUID generator, validator.
3. **Repository integration tests** (phase 4.1) — raw SQL queries against a real PostgreSQL container (testcontainers-go). Build tag: `integration`.
4. **Controller tests** (phase 4.2) — HTTP handlers via `httptest` with mock services.
5. **E2E tests** (phase 4.3) — full vertical slice (HTTP → DB) for auth flows. Build tag: `integration`.

**Test utilities:** `internal/testutil/db.go` (build tag: `integration`)
- `SetupTestDB(t)` — spins up PostgreSQL 16-alpine container, runs all migrations, returns `*sql.DB`. Container cleaned up on `t.Cleanup`.
- `TruncateTables(t, db)` — truncates all application tables between tests, preserving schema.

## Environment Variables

```
DB_HOST, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE
JWT_SECRET
IMAGE_STORE_ADDR      (gRPC address for image service, default: localhost:50051)
PORT
UNSPLASH_ACCESS_KEY   (optional; used by download-images script and fake-data-creator.sh)
```

## Preflight

**Ecosystem**: Go
**Config**: go.mod, Makefile, .golangci.yml, .air.toml
**Status**: ready

| Category | Status | Command |
|----------|--------|---------|
| Build    | ready  | `make dev` (via air, go build) |
| Check    | ready  | `make code-check` (gofumpt + golangci-lint) |
| Test     | ready  | `make test` (gotestsum) |

**Blockers**: none
**Warnings**: none
