.PHONY: help install-tools code-check dev docker-up docker-down docker-build db-up db-down db-remove test test-security wait-db start inject-data
.DEFAULT_GOAL := help

help:
	@echo ""
	@echo "  🛠️  Development:"
	@echo "    install-tools      - Install Go tools (gofumpt, golangci-lint, air, gotestsum)"
	@echo "    code-check         - Format and lint code"
	@echo "    dev                - Start application with databases"
	@echo "    start              - Interactive start (dev/docker/prod)"
	@echo "    test               - Run tests"
	@echo "    test-security      - Start docker containers and run security tests agains the api"
	@echo ""
	@echo "  🐳 Docker:"
	@echo "    docker-up          - Start all services"
	@echo "    docker-down        - Stop services"
	@echo "    docker-build       - Build API image"
	@echo ""
	@echo "  🗄️  Database:"
	@echo "    db-up              - Start databases"
	@echo "    db-down            - Stop databases"
	@echo "    db-remove          - Remove databases and volumes"
	@echo "    wait-db            - Wait for databases to be ready"
	@echo ""
	@echo "  🌱 Data:"
	@echo "    inject-data        - Seed fake stores and products into the running app"

install-tools:
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.7.2
	go install github.com/air-verse/air@latest
	go install gotest.tools/gotestsum@latest

code-check:
	gofumpt -l -w .
	golangci-lint run --fix ./...

wait-db:
	@echo "⏳ Waiting for databases..."
	@timeout 60 sh -c 'until docker exec lauti-market-postgres pg_isready -U postgres -d lauti_market > /dev/null 2>&1; do sleep 1; done' || (echo "❌ PostgreSQL timeout"; exit 1)
	@timeout 30 sh -c 'until docker exec lauti-market-redis redis-cli ping > /dev/null 2>&1; do sleep 1; done' || (echo "❌ Redis timeout"; exit 1)
	@echo "✅ Databases ready"

dev: db-up wait-db
	air -c .air.toml

start:
	@bash scripts/setupEnv.sh
	@echo ""
	@echo "🚀 Select how to start the project:"
	@echo "  1) Dev mode      - Go + air, local dbs, uses .env"
	@echo "  2) Docker (test) - Full stack in Docker, uses .env.test"
	@echo "  3) Prod mode     - Optimized Docker stack, uses .env.prod"
	@echo ""
	@read -p "Enter choice [1-3]: " choice; \
	case $$choice in \
		1) \
			echo ""; \
			echo "▶ Starting in dev mode with air (logs below)..."; \
			$(MAKE) dev || (echo "Dev failed, installing tools and retrying..."; $(MAKE) install-tools; $(MAKE) dev); \
			;; \
		2) \
			echo ""; \
			echo "▶ Starting Docker stack for testing (showing logs)..."; \
			[ -f .env.test ] || (echo ".env.test not found"; exit 1); \
			docker compose --env-file .env.test up; \
			;; \
		3) \
			echo ""; \
			echo "▶ Starting in production mode (optimized, showing logs)..."; \
			[ -f .env.prod ] || (echo ".env.prod not found"; exit 1); \
			docker compose --env-file .env.prod up; \
			;; \
		*) \
			echo "Invalid choice. Aborting."; \
			exit 1; \
			;; \
	esac

docker-up:
	@[ -f .env ] || (echo ".env not found"; exit 1)
	docker compose --env-file .env up -d

docker-down:
	docker compose down

docker-build:
	@[ -f .env ] || (echo ".env not found"; exit 1)
	docker compose build api

db-up:
	docker compose -f docker-compose.db.yml up -d

db-down:
	docker compose -f docker-compose.db.yml stop

db-remove:
	docker compose -f docker-compose.db.yml down -v

inject-data:
	go run ./cmd/seed

test:
	gotestsum --format=short-verbose

test-security:
	@[ -f .env ] || (echo ".env not found"; exit 1)
	@echo "🐳 Starting containers..."
	docker compose --env-file .env up -d --build
	@echo "⏳ Waiting for databases..."
	@timeout 60 sh -c 'until docker exec lauti-market-postgres pg_isready -U postgres > /dev/null 2>&1; do sleep 2; done' || (echo "❌ DB timeout"; docker compose logs; exit 1)
	@echo "⏳ Waiting for API..."
	@timeout 120 sh -c 'until curl -sf http://localhost:8080/health > /dev/null 2>&1; do sleep 2; done' || (echo "❌ API timeout"; docker compose logs; exit 1)
	@echo "✅ All services ready"
	@echo "🛡️ Running ASVS security tests..."
	go run ./cmd/securitytest
	@echo "🧹 Stopping containers..."
	docker compose down
