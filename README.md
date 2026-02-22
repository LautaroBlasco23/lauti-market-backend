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
