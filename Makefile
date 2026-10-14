AIR := $(shell which air 2>/dev/null || echo $(HOME)/go/bin/air)

.PHONY: help watch run test test-concurrency ci-local docker-up docker-down docker-build up down build web web-build

help:
	@echo "Available commands:"
	@echo "  make watch            - Run Go backend server with live-reloading (Air)"
	@echo "  make run              - Run Go backend server directly (:8080)"
	@echo "  make test             - Run all test suites"
	@echo "  make test-concurrency - Run concurrency & race-condition test"
	@echo "  make ci-local         - Run full local CI (tests + frontend build)"
	@echo "  make build            - Compile Go backend binary (bin/server)"
	@echo "  make web              - Run React + Vite frontend dev server (:3000)"
	@echo "  make web-build        - Build React production bundle"
	@echo "  make up               - Build and start full stack via Docker Compose"
	@echo "  make down             - Stop all Docker Compose services"
	@echo "  make docker-build     - Build Docker images"
	@echo "  make docker-up        - Start PostgreSQL container only"
	@echo "  make docker-down      - Stop all containers"

watch:
	@$(AIR) -c .air.toml

run:
	go run cmd/server/main.go

test:
	go test -v ./...

test-concurrency:
	go test -v -race tests/concurrency_test.go

ci-local: test test-concurrency web-build
	@echo "All local CI checks passed successfully."

build:
	@mkdir -p bin
	go build -ldflags="-s -w" -o bin/server cmd/server/main.go
	@echo "Backend binary built at bin/server"

web:
	cd web && npm run dev

web-build:
	cd web && npm install && npm run build

up:
	docker compose up --build -d

down:
	docker compose down

docker-build:
	docker compose build

docker-up:
	docker compose up -d postgres

docker-down:
	docker compose down


