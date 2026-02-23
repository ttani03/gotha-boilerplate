.PHONY: generate build run dev test lint css css-watch docker-up docker-down migrate-up migrate-down migrate-create tilt-up

# =============================================================================
# Code Generation
# =============================================================================

## Generate templ and sqlc code
generate:
	templ generate
	sqlc generate

# =============================================================================
# Build & Run
# =============================================================================

## Build the Go binary
build: generate css
	go build -o ./tmp/main ./cmd/server

## Run the application locally
run: build
	./tmp/main

## Start development with hot reload (air)
dev: css
	docker compose -f docker/docker-compose.yml up -d db
	@echo "Waiting for database to be ready..."
	@sleep 3
	air -c .air.toml

# =============================================================================
# Testing & Linting
# =============================================================================

## Run all tests
test:
	go test ./... -v -count=1

## Run linter
lint:
	golangci-lint run

# =============================================================================
# Frontend (CSS)
# =============================================================================

## Build Tailwind CSS
css:
	npm run css:build

## Watch Tailwind CSS for changes
css-watch:
	npm run css:watch

# =============================================================================
# Docker
# =============================================================================

## Start all services with Docker Compose
docker-up:
	docker compose -f docker/docker-compose.yml up --build

## Stop all services
docker-down:
	docker compose -f docker/docker-compose.yml down

# =============================================================================
# Database Migrations (goose)
# =============================================================================

MIGRATIONS_DIR := internal/db/migrations
DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/gotha?sslmode=disable

## Apply all pending migrations
migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

## Rollback the last migration
migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

## Create a new migration file
migrate-create:
	@read -p "Migration name: " name; \
	goose -dir $(MIGRATIONS_DIR) create $$name sql

# =============================================================================
# Tilt (Development)
# =============================================================================

## Start Tilt for development
tilt-up:
	tilt up

# =============================================================================
# Setup (Initial)
# =============================================================================

## Install all dependencies
setup:
	go mod download
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/air-verse/air@latest
	mise trust
	mise install
	npm install

# =============================================================================
# Help
# =============================================================================

## Show this help
help:
	@echo "Available targets:"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/^## /  /'
	@echo ""
	@grep -E '^[a-zA-Z_-]+:' Makefile | sed 's/:.*//' | sort | sed 's/^/  make /'
