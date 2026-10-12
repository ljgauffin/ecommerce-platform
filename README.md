# E-Commerce Platform API

A e-commerce platform built with Go (1.21+), PostgreSQL, and Clean Architecture for GilaSW interview

## Quick Start

```bash
# 1. Start PostgreSQL (on port 5433 to avoid local conflicts with any instance installed)
make docker-up

# 2. Run server with live reload
make watch

# 3. Swagger UI documentation
open http://localhost:8080/swagger
```

---

## CSV Ingestion (`POST /api/v1/products/import-csv`)

The endpoint accepts `multipart/form-data` with a `file` field

### Key Rules & Behavior

1. **Delimiters & Formatting**:
   - Detects comma (`,`) or tab (`\t`).
   - Normalizes Windows double carriage return (`\r\r\n`) 

2. **Duplicate SKUs in File**:
   - If an SKU appears more than once in the same CSV, all occurrences are rejected and reported in the `errors` list. Could have keeped last appereance, but if the duplication was a mistake, should be the user who decides which appearance to keep.

3. **Response Breakdown**:
   - The endpoint respondes the result of the process:
     ```json
     {
       "total_rows": 97,
       "created": 0,
       "updated": 84,
       "imported": 84,
       "failed": 13,
       "errors": [ ... ]
     }
     ```
   - `created`: newly inserted products.
   - `updated`: existing products updated.
   - `failed`: rows rejected due to restrictions previusly mentioned.

4. **Security & Content Sanitization (XSS, HTML tags & scripts)**:
   - I found injection attempts in the dataset such as (`<script>alert('xss')</script>`, HTML tags, and SQL injection strings).
   - **Backend storage level**: `<script>` blocks and arbitrary HTML tags are stripped during sanitization. It's trimmed and normalized. SQL injection is managed by the GO ORM.
   - **Frontend level**: Escapes string content in JSX.

5. **Data Normalization**:
   - Prices like `"$29.99"` or `"free"` are normalized to `29.99` and `0.0`.
   - Negative numbers for stock or weight are rejected.

---

## Checkout & Concurrency (`POST /api/v1/orders/checkout`)

- **Row-level locking**: Used SELECT ... FOR UPDATE inside GORM transactions to guarantee stock deduction and prevent database race conditions over high application demand.
- **Idempotency**: `Idempotency-Key` header added to avoid duplicate or replayed requests and return the cached order without deducting stock twice.
- **Automated test**: `tests/concurrency_test.go` simulates 20 simultaneous buyers trying to buy 5 items of stock.

---

## Running Tests

```bash
make test
```
