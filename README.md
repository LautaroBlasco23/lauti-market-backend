# Lauti Market Backend

Backend API for my marketplace personal project.

## Architecture

```
internal/
├── api/                    # Shared infrastructure
└── {module}/               # auth, user, store, product, order, image, payment
    ├── domain/             # Business rules, entities, repository interfaces
    ├── application/        # Use cases
    └── infrastructure/     # HTTP handlers, repositories, wiring
```

**Dependency direction:** `infrastructure → application → domain`

Each module exposes a `Wire()` function in `infrastructure/wiring.go` that initializes all components and registers routes.

**Module initialization order:** User → Store → Auth → Image → Product → Order → Payment

## Dependencies

- [golang-jwt/jwt](https://github.com/golang-jwt/jwt) - JWT tokens
- [google/uuid](https://github.com/google/uuid) - ID generation
- [x/crypto](https://pkg.go.dev/golang.org/x/crypto) - bcrypt password hashing
- [go-playground/validator](https://github.com/go-playground/validator) - Struct validation
- [godotenv](https://github.com/joho/godotenv) - Load .env files
- [mercadopago/sdk-go](https://github.com/mercadopago/sdk-go) - MercadoPago Checkout Pro integration
- [testify](https://github.com/stretchr/testify) - Test assertions and mocking
- [testcontainers-go](https://github.com/testcontainers/testcontainers-go) - PostgreSQL containers for integration tests

## Run

1. Copy the environment file:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

2. Start the server:
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

Server starts on `:8080` (Docker) or `PORT` env variable (default `:8000` in dev).

### Seed data

```bash
# Populate the database with fake stores and products
make inject-data
```

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

## ngrok

Expose your local backend to the internet with [ngrok](https://ngrok.com/).

### Quick start

1. Install ngrok:
   ```bash
   curl -s https://ngrok-agent.s3.amazonaws.com/ngrok.asc | sudo tee /etc/apt/trusted.gpg.d/ngrok.asc >/dev/null && echo "deb https://ngrok-agent.s3.amazonaws.com buster main" | sudo tee /etc/apt/sources.list.d/ngrok.list && sudo apt update && sudo apt install ngrok
   ```
   Or download from https://ngrok.com/download

2. Add your authtoken (once):
   ```bash
   ngrok config add-authtoken <YOUR_TOKEN>
   ```
   Get your token from https://dashboard.ngrok.com/get-started/your-authtoken

3. Start your backend and expose it:
   ```bash
   # Terminal 1: start the server
   make dev

   # Terminal 2: expose port 8080
   ngrok http 8080
   ```

4. Copy the `Forwarding` URL (e.g., `https://abc123.ngrok-free.app`)

### ngrok with Docker

If running the backend via `make docker-up`:
```bash
ngrok http 8080
```
