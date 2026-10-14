# E-Commerce Platform

A production-ready e-commerce platform built with Go (1.23+), PostgreSQL, and React + Vite using Clean Architecture.

> **Example CSV file downloaded on**: October 11, 2026

---

## How to Run Locally

### Option A: Full Docker Stack (Recommended)

Run PostgreSQL, the Go API, and the React web frontend in Docker containers:

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
make docker-up

# 2. Run server with live reload
make watch

# 3. In another terminal, run the frontend
make web
```

---

## Features & UI

The frontend is a decoupled single-page application built with React and Vite in `web/` (running on `:3000`):

- **Product CRUD**:
  - **Create**: "+ New Product" button opens a modal to create products (`POST /api/v1/products`).
  - **Read**: Catalogue grid with stock badges, price formatting, and pagination/filtering.
  - **Update**: Edit button on each product card to update details (`PUT /api/v1/products/:id`).
  - **Delete**: Delete button with confirmation prompt (`DELETE /api/v1/products/:id`).
- **Search & Filtering (Server-side)**:
  - Instant text search across product name, SKU, category, and description executed server-side via SQL query parameters.
  - Category filters dynamically populated from `GET /api/v1/products/categories`.
  - In-stock availability toggle and dynamic sorting (newest, price asc/desc, alphabetical) processed directly by the database endpoint.
  - Debounced client requests (250ms) to prevent server spam while preserving real-time UX.
- **Purchase & Checkout**:
  - Slide-out Cart Drawer with quantity steppers bounded by live stock.
  - Faked payment checkout with customer email validation.
  - Atomic stock deduction and `Idempotency-Key` header generation to prevent double purchases.
- **CSV Ingestion Modal**:
  - Drag-and-drop file upload.
  - Real-time processing summary (`total_rows`, `created`, `updated`, `failed`).
  - Audit table displaying row-by-row rejection reasons.

---

## Architectural Decisions & Approach

### 1. Clean Architecture
Codebase is structured into four decoupled layers:
- `internal/domain`: Pure entities (`Product`, `Category`, `Order`, `OrderItem`, `IdempotencyRecord`) and repository interfaces. No third-party dependencies.
- `internal/usecase`: Business logic (`CSVImportUseCase`, `CheckoutUseCase`).
- `internal/infrastructure`: PostgreSQL persistence using GORM and database migrations.
- `internal/presentation`: HTTP handlers (Chi router), input sanitization, and Swagger OpenAPI spec.

### 2. Concurrency & Inventory Protection
- During checkout, `SELECT ... FOR UPDATE` locks product rows within a single SQL transaction.
- Prevents race conditions and overselling when multiple buyers attempt to purchase the same inventory simultaneously.
- Verified with an automated test in `tests/concurrency_test.go` (20 concurrent buyers contending for 5 items).

### 3. Idempotency
- Checkout requests require an `Idempotency-Key`.
- Persisted in PostgreSQL with response caching. If a network retry occurs with the same key, the cached order is returned without deducting stock again.

### 4. Dedicated Categories Table & Controlled Entry
- Categories are managed as a first-class entity in their own database table (`categories`) with unique constraints.
- The `products` table stores solely `category_id` referencing `categories(id)` via foreign key relationship, removing redundant text columns from the product schema.
- In manual product creation/updating, arbitrary free-text entry is restricted: products must reference an existing category from the database, preventing fragmentation and typos.
- Categories have full RESTful management (`GET /api/v1/categories`, `POST /api/v1/categories`, `DELETE /api/v1/categories/:id` with foreign key protection).

### 5. CSV Ingestion & Security Rules
- **Preserved CSV Format**: The proposed CSV structure (`name, sku, description, category, price, stock, weight_kg`) is strictly preserved without any format changes.
- **Automatic Category Creation**: When importing products from a CSV, if a row contains a category that does not yet exist in the `categories` table, it is automatically created and linked to the product on-the-fly, ensuring frictionless bulk ingestion without manual pre-configuration.
- **Delimiter Detection**: Handles both comma (`,`) and tab (`\t`) separated files, as well as Windows double carriage returns (`\r\r\n`).
- **Duplicate SKU Policy**: If an SKU appears more than once in the same CSV file, all occurrences are rejected and reported in the row errors array. If duplication was an accidental mistake, the user should decide which row is correct rather than silently keeping the last occurrence.
- **Content Handling & Security**: Text fields preserve exact raw characters (including quotes, special characters, and tags) with only leading/trailing whitespace trimmed. Backend HTML tag stripping and HTML entity escaping were deliberately removed to avoid corrupting data (e.g. quotes turning into `&#34;` or apostrophes into `&#39;`). React naturally sanitizes JSX text expressions to prevent client-side XSS, and parameterized SQL queries in the persistence layer prevent SQL injection.
- **Data Normalization**: Values like `"$29.99"` or `"free"` are normalized to numeric decimals. Negative values for stock or weight are rejected.

---

## Alternatives Considered

- **Dedicated Categories Table vs Free-Text Strings**: Storing categories as unconstrained strings allows typos, inconsistent casings, and data pollution. Promoting categories to their own table guarantees relational integrity. To prevent this from breaking automated CSV imports, non-existent categories in the CSV are dynamically provisioned in the database during ingestion.
- **PostgreSQL vs NoSQL (MongoDB)**: Considered NoSQL for flexible product attributes, but chose PostgreSQL because strict ACID transactions and row-level locking (`SELECT FOR UPDATE`) are essential for preventing inventory overselling during concurrent checkouts.
- **GORM vs Raw SQL (pgx)**: Evaluated raw SQL for performance, but chose GORM for clean repository interfaces and automated migrations, using raw SQL locks where critical concurrency control is needed.
- **Single-binary embed vs Decoupled Frontend**: Initially evaluated compiling React into the Go binary (`go:embed`), but pivoted to a decoupled setup (React + Vite on `:3000`, Go API on `:8080`) to allow faster frontend hot-module reloading and independent deployments.
- **Duplicate SKU: Last-wins vs Complete Rejection**: Considered keeping the last occurrence of duplicate SKUs in a CSV file, but decided on complete rejection so the user is alerted to the ambiguity instead of silently overwriting data.
- **HTML Tag Stripping & Entity Escaping vs Raw String Preservation**: Evaluated stripping `<script>`/HTML tags and applying `html.EscapeString`, but discarded it because it corrupted legitimate characters (quotes became `&#34;` and apostrophes became `&#39;`), whereas React's virtual DOM inherently guards against script execution in JSX strings without mutating stored data.

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
- Spins up a real PostgreSQL 16 service container.
- Runs all unit & integration tests with the Go race detector (`-race`).
- Executes the concurrency benchmark (`tests/concurrency_test.go`).
- Builds the backend binary (`bin/server`).
- Builds the React production bundle (`web/dist`).
- Validates Docker Compose image builds.

---

## Makefile Command Reference

| Command | Description |
|---|---|
| `make up` | Start full platform (Postgres, Go API, React Web) via Docker Compose |
| `make down` | Stop all Docker Compose containers |
| `make watch` | Run Go backend with live reloading (`air`) on `:8080` |
| `make web` | Run React + Vite frontend dev server on `:3000` |
| `make test` | Run all test suites |
| `make test-concurrency` | Run race-condition checkout benchmark |
| `make ci-local` | Run complete CI suite locally |
| `make build` | Compile Go binary to `bin/server` |
| `make web-build` | Install frontend dependencies and build production assets |
| `make docker-build` | Build Docker images for both backend and frontend |
| `make docker-up` | Start PostgreSQL container only (port `5433`) |
| `make docker-down` | Stop PostgreSQL container |

---

## REST API Reference

All endpoints are documented interactively via Swagger UI at `/swagger`.

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/v1/products` | List products with pagination, server-side search, and category/price filters |
| `POST` | `/api/v1/products` | Create a new product (validates category in `categories` table) |
| `GET` | `/api/v1/products/:id` | Get product details by UUID |
| `PUT` | `/api/v1/products/:id` | Update an existing product |
| `DELETE` | `/api/v1/products/:id` | Delete a product |
| `POST` | `/api/v1/products/import-csv` | Bulk import products via multipart CSV upload |
| `GET` | `/api/v1/products/categories` | List distinct category names (string array) |
| `GET` | `/api/v1/categories` | List all category objects (`id`, `name`, timestamps) |
| `POST` | `/api/v1/categories` | Create a new category |
| `DELETE` | `/api/v1/categories/:id` | Delete a category (prevented if products are associated) |
| `POST` | `/api/v1/orders/checkout` | Process order checkout (`Idempotency-Key` header required) |
| `GET` | `/api/v1/orders/:id` | Retrieve order by ID |
| `GET` | `/swagger` | Interactive OpenAPI documentation UI |
| `GET` | `/swagger/doc.json` | OpenAPI 3.0 JSON specification |
| `GET` | `/health/live` | Liveness health check |
| `GET` | `/health/ready` | Readiness health check (validates database connection) |

