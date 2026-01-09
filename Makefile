.PHONY: help install-tools code-check dev docker-up docker-down docker-build db-up db-down db-remove test test-security
.DEFAULT_GOAL := help

help:
	@echo ""
	@echo "  🛠️  Development:"
	@echo "    install-tools      - Install Go tools (gofumpt, golangci-lint, air, gotestsum)"
	@echo "    code-check         - Format and lint code"
	@echo "    dev                - Start application with databases"
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

install-tools:
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.7.2
	go install github.com/air-verse/air@latest
	go install gotest.tools/gotestsum@latest

code-check:
	gofumpt -l -w .
	golangci-lint run --fix ./...

dev: db-up
	ENV_FILE=.env air -c .air.toml

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

test:
	gotestsum --format=short-verbose

test-security:
	@[ -f .env ] || (echo ".env not found"; exit 1)
	@echo "🐳 Starting containers..."
	docker compose --env-file .env up -d --build

	@echo "⏳ Waiting for containers to be healthy..."
	@sh -c '\
		set -e; \
		for i in $$(seq 1 60); do \
			unhealthy=$$(docker inspect --format="{{.State.Health.Status}}" $$(docker compose ps -q) 2>/dev/null | grep -v healthy || true); \
			if [ -z "$$unhealthy" ]; then \
				echo "✅ All containers healthy"; \
				exit 0; \
			fi; \
			sleep 2; \
		done; \
		echo "❌ Containers not healthy in time"; \
		docker compose logs; \
		exit 1'

	@echo "🛡️ Running ASVS security tests..."
	go run ./cmd/securitytest

	@echo "🧹 Stopping containers..."
	docker compose down
