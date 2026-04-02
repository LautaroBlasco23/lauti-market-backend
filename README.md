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
- [joho/godotenv](https://github.com/joho/godotenv) - `.env` loading

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

Server starts on `:8080` (Docker) or the `PORT` env variable (default `:8000` in dev).

### Seed fake data

With the server running, inject sample stores and products:

```bash
make inject-data
```

Optionally set `UNSPLASH_ACCESS_KEY` in `.env` to attach real product images via the Unsplash API. A random suffix is appended to every run so it is safe to call multiple times.

## CI/CD

The GitHub Actions pipeline builds, tests, and pushes a Docker image to Docker Hub on every push to `main`.

Required repository secrets (`Settings → Secrets and variables → Actions`):

| Secret               | Description                                                                 |
|----------------------|-----------------------------------------------------------------------------|
| `DOCKERHUB_USERNAME` | Your Docker Hub username                                                    |
| `DOCKERHUB_TOKEN`    | Docker Hub access token — generate at hub.docker.com → Account Settings → Security |
