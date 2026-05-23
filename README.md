# Go Hiring Challenge

This repository contains a Go application for managing products and their prices, including functionalities for CRUD operations and seeding the database with initial data.

## Project Structure

1. **cmd/**: Contains the main application and seed command entry points.

   - `server/main.go`: The main application entry point, serves the REST API.
   - `seed/main.go`: Command to seed the database with initial product data.

2. **app/**: Contains the application logic.
3. **sql/**: Contains a very simple database migration scripts setup.
4. **models/**: Contains the data models and repositories used in the application.
5. `.env`: Environment variables file for configuration.

## Setup Code Repository

1. Create a github/bitbucket/gitlab repository and push all this code as-is.
2. Create a new branch, and provide a pull-request against the main branch with your changes. Instructions to follow.

## Application Setup

- Ensure you have Go installed on your machine.
- Ensure you have Docker installed on your machine.
- Important makefile targets:
  - `make tidy`: will install all dependencies.
  - `make docker-up`: will start the required infrastructure services via docker containers.
  - `make seed`: ⚠️ Will destroy and re-create the database tables.
  - `make test`: Will run the tests.
  - `make run`: Will start the application.
  - `make docker-down`: Will stop the docker containers.

Follow up for the assignemnt here: [ASSIGNMENT.md](ASSIGNMENT.md)

## API Endpoints

Base URL: `http://localhost:8484` (configured via `HTTP_PORT` in `.env`)

---

### `GET /health`

```bash
curl http://localhost:8484/health
```

```json
{"status": "ok"}
```

---

### `GET /catalog`

Returns a paginated list of products. Supports filtering by category and price.

| Query param | Type   | Default | Description                        |
|-------------|--------|---------|------------------------------------|
| `category`  | string | —       | Filter by category code            |
| `max_price` | number | —       | Filter products with price less than this value |
| `offset`    | int    | `0`     | Pagination offset                  |
| `limit`     | int    | `10`    | Pagination limit (max `100`)       |

```bash
curl "http://localhost:8484/catalog?category=clothing&max_price=20&offset=0&limit=5"
```

```json
{
  "products": [
    {
      "code": "PROD001",
      "price": 10.99,
      "category": { "code": "clothing", "name": "Clothing" }
    }
  ],
  "total": 1,
  "offset": 0,
  "limit": 5
}
```

| Status | Reason                                        |
|--------|-----------------------------------------------|
| `200`  | Success                                       |
| `400`  | Invalid `offset`, `limit`, or `max_price`     |
| `500`  | Database error                                |

---

### `GET /catalog/{code}`

Returns a single product with its variants. Variant price falls back to the product price when not set.

```bash
curl http://localhost:8484/catalog/PROD001
```

```json
{
  "code": "PROD001",
  "price": 10.99,
  "category": { "code": "clothing", "name": "Clothing" },
  "variants": [
    { "name": "Variant A", "sku": "SKU001A", "price": 11.99 },
    { "name": "Variant B", "sku": "SKU001B", "price": 10.99 },
    { "name": "Variant C", "sku": "SKU001C", "price": 10.99 }
  ]
}
```

| Status | Reason          |
|--------|-----------------|
| `200`  | Success         |
| `404`  | Product not found |
| `500`  | Database error  |

---

### `GET /categories`

Returns all available categories.

```bash
curl http://localhost:8484/categories
```

```json
{
  "categories": [
    { "code": "clothing",    "name": "Clothing"     },
    { "code": "shoes",       "name": "Shoes"        },
    { "code": "accessories", "name": "Accessories"  }
  ]
}
```

| Status | Reason         |
|--------|----------------|
| `200`  | Success        |
| `500`  | Database error |

---

### `POST /categories`

Creates a new category. Requires `Content-Type: application/json`.

```bash
curl -X POST http://localhost:8484/categories \
  -H "Content-Type: application/json" \
  -d '{"code": "bags", "name": "Bags"}'
```

```json
{"code": "bags", "name": "Bags"}
```

| Status | Reason                               |
|--------|--------------------------------------|
| `201`  | Category created                     |
| `400`  | Missing or empty `code` / `name`     |
| `409`  | Category code already exists         |
| `415`  | Content-Type is not application/json |
| `500`  | Database error                       |
