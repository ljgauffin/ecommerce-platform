AIR := $(shell which air 2>/dev/null || echo $(HOME)/go/bin/air)

.PHONY: help db db-down api api-build web web-build watch run test test-concurrency ci-local up down build docker-up docker-down docker-build

help:
	@echo "Available commands:"
	@echo "  make db               - Start PostgreSQL container (port 5433)"
	@echo "  make db-down          - Stop PostgreSQL container"
	@echo "  make api              - Run Go backend server with live-reloading (Air on :8080)"
	@echo "  make web              - Run React + Vite frontend dev server (:3000)"
	@echo "  make api-build        - Compile Go backend binary (bin/server)"
	@echo "  make web-build        - Build React production bundle"
	@echo "  make test             - Run all test suites"
	@echo "  make test-concurrency - Run concurrency & race-condition test"
	@echo "  make ci-local         - Run full local CI (tests + frontend build)"
	@echo "  make up               - Build and start full stack via Docker Compose"
	@echo "  make down             - Stop all Docker Compose services"

db:
	docker compose up -d postgres

db-down:
	docker compose stop postgres

api:
	@$(AIR) -c .air.toml

api-build:
	@mkdir -p bin
	go build -ldflags="-s -w" -o bin/server cmd/server/main.go
	@echo "Backend binary built at bin/server"

web:
	cd web && npm run dev

web-build:
	cd web && npm install && npm run build

run:
	go run cmd/server/main.go

test:
	go test -v ./...

test-concurrency:
	go test -v -race tests/concurrency_test.go

ci-local: test test-concurrency web-build
	@echo "All local CI checks passed successfully."

up:
	docker compose up --build -d

down:
	docker compose down

docker-build:
	docker compose build

watch: api
build: api-build
docker-up: db
docker-down: db-down


