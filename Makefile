AIR := $(shell which air 2>/dev/null || echo $(HOME)/go/bin/air)

.PHONY: help watch run test docker-up docker-down build web web-build

help:
	@echo "Available commands:"
	@echo "  make watch       - Run Go backend server with live-reloading (Air)"
	@echo "  make run         - Run Go backend server directly (:8080)"
	@echo "  make test        - Run all test suites"
	@echo "  make build       - Compile Go backend binary (bin/server)"
	@echo "  make web         - Run React + Vite frontend dev server (:3000)"
	@echo "  make web-build   - Build React production bundle"
	@echo "  make docker-up   - Start PostgreSQL container"
	@echo "  make docker-down - Stop PostgreSQL container"

watch:
	@$(AIR) -c .air.toml

run:
	go run cmd/server/main.go

test:
	go test -v ./...

build:
	@mkdir -p bin
	go build -ldflags="-s -w" -o bin/server cmd/server/main.go
	@echo "Backend binary built at bin/server"

web:
	cd web && npm run dev

web-build:
	cd web && npm install && npm run build

docker-up:
	docker compose up -d postgres

docker-down:
	docker compose down

