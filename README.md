# Lauti Market Backend

Backend API for my marketplace personal project.

## Architecture

```
internal/
├── api/                    # Shared infrastructure
└── {entity}/
    ├── domain/             # Business rules, entities, repository interfaces
    ├── application/        # Use cases
    └── infrastructure/     # HTTP handlers, repositories, wiring
```

**Dependency direction:** `infrastructure → application → domain`

Each module exposes a `Wire()` function in `infrastructure/wiring.go` that initializes all components and registers routes.

## Dependencies

- [golang-jwt/jwt](https://github.com/golang-jwt/jwt) - JWT tokens
- [google/uuid](https://github.com/google/uuid) - ID generation
- [x/crypto](https://pkg.go.dev/golang.org/x/crypto) - bcrypt password hashing
- [testify](https://github.com/stretchr/testify) - Test assertions and mocking
- [testcontainers-go](https://github.com/testcontainers/testcontainers-go) - PostgreSQL containers for integration tests

## Run

```bash
go mod tidy
make db-up
make dev
```

or

```bash
make install-tools
make db-up
make dev
```

Server starts on `:8080`.

## Testing

```bash
# Unit + controller tests (no Docker required)
make test

# Include integration and E2E tests (requires Docker)
go test -tags=integration ./...

# Coverage report (opens browser)
make test-coverage

# ASVS security tests (requires running Docker stack)
make test-security
```

Integration and E2E tests use the `//go:build integration` tag and spin up a real PostgreSQL 16-alpine container via testcontainers-go. Migrations are applied automatically.
