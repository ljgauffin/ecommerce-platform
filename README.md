# E-Commerce Platform

An e-commerce platform built with Go, PostgreSQL, and React + Vite using Clean Architecture for a GilaSW interview. 
In the project I tried to showcase my knowledge on some of the requeried skills. 
As a sidenote. I've decided to use GO for the API mainly as it was one of the mentioned languages in the role description. Even though I have worked with it in the past, it's been a while since I've used it. Lately my main focused has been TS in NodeJS with NestJS. 

> **Example CSV file downloaded on**: October 11, 2026

---

## How to Run Locally

### Option A: Full Docker Stack

Runs PostgreSQL, the Go API, and the React web frontend in Docker containers:

```bash
docker compose up --build
# or
make up
```

- **Frontend Web UI**: `http://localhost:3000`
- **Backend API Server**: `http://localhost:8080`
- **Swagger OpenAPI Documentation**: `http://localhost:8080/swagger`

To stop all containers:
```bash
docker compose down
# or
make down
```

### Option B: Local Development with Live Reload

```bash
# 1. Start PostgreSQL (mapped to port 5433 to avoid local conflicts)
make db

# 2. Run API with live reload
make api

# 3. In another terminal, run the frontend
make web
```

---

## Features & UI

The frontend is an SPA application in React and Vite in `web/` (running on `:3000`):

- **Product CRUD**: Creation, editing, deletion, and pagination.
- **Search & Filtering**: Server-side text search (name, SKU, description), category selector, stock status, and sorting.
- **Storefront**: Product catalog with infinite scrolling, cart drawer, and checkout with idempotency.
- **CSV Import**: Drag-and-drop bulk ingestion with row-by-row validation and error summary.

---

## Architectural Decisions

### 1. Clean Architecture
Codebase is structured into four layers:
- `internal/domain`: Extrictly entities (`Product`, `Category`, `Order`, `OrderItem`, `IdempotencyRecord`) and repository interfaces.
- `internal/usecase`: Business logic (`CSVImportUseCase`, `CheckoutUseCase`).
- `internal/infrastructure`: PostgreSQL persistence using GORM and database migrations.
- `internal/presentation`: HTTP handlers, input sanitization, and Swagger OpenAPI spec.

### 2. Concurrency & Inventory Protection
- During checkout, `SELECT ... FOR UPDATE` to lock product rows within a single SQL transaction to prevent race conditions
- Verified with an automated test in `tests/concurrency_test.go` (20 concurrent buyers trying to buy 5 items).

### 3. Idempotency
- Checkout requests require an `Idempotency-Key`. If a network retry occurs with the same key, the cached order is returned without deducting stock again.

### 4. Dedicated Categories Table & Controlled Entry
- Categories have full RESTful management. Didnt make a category CRUD in frontend. But csv ingestation creates new categories automatically.
- Manual CRUD of products requires to select existing categories.

### 5. CSV Ingestion & Security Rules
- **Preserved CSV Format**: The proposed CSV structure (`name, sku, description, category, price, stock, weight_kg`) is preserved.
- **Automatic Category Creation**: When importing products from a CSV, if a row contains a category that does not yet exist in the `categories` table, it is created.
- **Delimiter Detection**: Handles common CSV delimiters, as well as Windows double carriage returns (`\r\r\n`).
- **Duplicate SKU Policy**: If an SKU appears more than once in the same CSV file, all occurrences are rejected and reported in the row errors array (and frontend result modal), so the user can decide which product is correct.
- **Content Handling & Security**: Text fields preserve raw characters (including quotes, special characters, and tags) with only leading/trailing whitespace trimmed. I've seen xml and sql injection, as well as double quote in titles; react naturally sanitizes JSX text expressions to prevent client-side XSS, and parameterized SQL queries in the persistence layer prevent SQL injection.
- **Data Normalization**: Values like `"$29.99"` or `"free"` are normalized to numeric decimals. Negative values for stock or weight are rejected.

---

## Alternatives Considered

- **SQL vs NoSQL (MongoDB)**: Considered NoSQL for flexible product attributes, but chose PostgreSQL for row level locking considering a checkout as a sensible transaction.
- **GORM**: Evaluated raw SQL for performance, but chose GORM for clean repository interfaces and automated migrations, using raw SQL locks where critical concurrency control is needed. Also controls SQL injection.
- **Single-binary embed vs Decoupled Frontend**: Initially evaluated compiling React into the Go binary (`go:embed`), but changed to a decoupled setup (React + Vite on `:3000`, Go API on `:8080`) mainly because is a better known setup to me.
- **Duplicate SKU: Last-wins vs Complete Rejection**: Considered keeping the last occurrence of duplicate SKUs in a CSV file, but decided on complete rejection so the user can decide best course of action.
- **HTML Tag Stripping & Entity Escaping vs Raw String Preservation**: Evaluated stripping `<script>`/HTML tags and applying `html.EscapeString`, but discarded it would make the titles empty and end up rejecting the row, and React manages natively in JSX.

---

## Production & Scaling Considerations

To scale this service in a cloud environment (e.g., AWS/Kubernetes):
- **Event-Driven Architecture**: Decouple post-checkout operations (order fulfillment, email notifications, analytics) by publishing `OrderPlaced` and `InventoryDeducted` events to Kafka or similar.
- **Read/Write Segregation & Caching**: Introduce Redis for caching repeted product catalog queries, and routing read queries to PostgreSQL read replicas and write traffic (inventory updates, checkout) to the primary database.
- **Distributed Rate Limiting**: Complement row-level database locks with Redis-based distributed rate limiters to shield the checkout API during flash sales.
- **Observability**: Instrument endpoints with OpenTelemetry for distributed tracing, Prometheus metrics (`http_request_duration_seconds`, `db_connection_pool_stats`), and structured JSON logging.

The improvements mentioned previusly were not implemented in order to avoid expanding the scope of the project too much and due to a lack of time. 

---

## Testing & CI/CD

### Running Tests Locally

```bash
# Run all unit and integration test suites
make test

# Run the live concurrency benchmark with race detector
make test-concurrency

# Run full local CI (backend tests, race detector, frontend build)
make ci-local
```

### GitHub Actions CI Pipeline

The project includes an automated CI workflow at `.github/workflows/ci.yml` that executes on every push and pull request to `main`:
- Spins up a  PostgreSQL service container.
- Runs all unit & integration tests with the Go race detector (`-race`).
- Executes the concurrency benchmark (`tests/concurrency_test.go`).
- Builds the backend binary (`bin/server`).
- Builds the React production bundle (`web/dist`).
- Validates Docker Compose image builds.

---

## Makefile

Run `make help` to see the list of all available commands and their descriptions:

```bash
make help
```

---

## REST API Reference

All endpoints are documented interactively via Swagger UI at `/swagger`.



