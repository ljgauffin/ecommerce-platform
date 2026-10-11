AIR := $(shell which air 2>/dev/null || echo $(HOME)/go/bin/air)

.PHONY: help watch run test docker-up docker-down

help:
	@echo "Available commands:"
	@echo "  make watch       - Run server with live-reloading (Air)"
	@echo "  make run         - Run server directly"
	@echo "  make test        - Run all test suites"
	@echo "  make docker-up   - Start PostgreSQL container"
	@echo "  make docker-down - Stop PostgreSQL container"

watch:
	@$(AIR) -c .air.toml

run:
	go run cmd/server/main.go

test:
	go test -v ./...

docker-up:
	docker compose up -d postgres

docker-down:
	docker compose down
