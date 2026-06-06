# Products Service

`products-service` is a Go-based gRPC microservice for managing catalog products in my "runaway e-commerce application".

It provides product CRUD operations, total-price calculation for order items, Redis-backed product caching, Kafka-driven stock reduction, and product embedding storage backed by PostgreSQL/pgvector.

## What it does

- Exposes a gRPC API for product operations
- Stores product data in PostgreSQL via GORM
- Caches single-product lookups in Redis
- Consumes `orders.placed` Kafka events to decrease stock
- Generates Gemini embeddings for newly created products
- Stores embedding vectors in configured pgvector tables
- Protects write operations with JWT-based gRPC auth interceptors
- Logs unary and streaming gRPC requests with Zap

## Tech stack

- Go `1.25`
- gRPC / Protocol Buffers
- PostgreSQL + GORM
- Redis
- Kafka (`segmentio/kafka-go`)
- Google GenAI (`gemini-embedding-2-preview`)
- JWT auth
- Docker

## Project structure

```text
products/
├── api/                 # gRPC server setup and interceptor wiring
├── cache/               # Redis client setup
├── db/                  # PostgreSQL connection and migrations
├── gen/pb/              # Generated protobuf/go gRPC code
├── handlers/            # gRPC handlers
├── interceptors/        # Auth and logging interceptors
├── internal/
│   ├── auth/            # JWT validation and claims
│   ├── config/          # Environment-based configuration
│   └── kafka/           # Kafka consumer and topics
├── models/              # Product and embedding models
├── services/            # Business logic
├── store/               # Postgres store implementation
├── Dockerfile
├── go.mod
├── main.go
└── run.sh
```

## gRPC API

Service: `products.ProductsService`

### RPC methods

| Method | Type | Auth | Notes |
|---|---|---:|---|
| `GetProductByID` | Unary | No | Reads product ID from gRPC metadata key `product_id` |
| `GetProductsByIDs` | Server streaming | No | Streams products for the provided list of IDs |
| `AddProduct` | Unary | Yes | Requires a valid bearer token with seller role |
| `CalculateTotalPrice` | Unary | Yes | Calculates the total using stored product prices |
| `UpdateProduct` | Unary | Yes | Requires seller role |
| `DeleteProduct` | Unary | Yes | Requires seller role |

## Request/response notes

### `GetProductByID`
This RPC currently accepts `google.protobuf.Empty` as the request body. The product ID must be sent as gRPC metadata:

- Metadata key: `product_id`
- Example value: `1`

### `GetProductsByIDs`
Accepts a request with:

```json
{
  "product_ids": [1, 2, 3]
}
```

and returns a server stream of product records.

### `AddProduct`
Request shape:

```json
{
  "name": "Wireless Mouse",
  "price": 29.99,
  "image": "https://example.com/mouse.png",
  "category": "Electronics",
  "description": "Compact wireless mouse",
  "stock": 20,
  "in_stock": true,
  "tags": ["wireless", "mouse", "usb"]
}
```

On success, the service:

1. saves the product,
2. caches it in Redis,
3. generates an embedding via Gemini,
4. stores the embedding in the configured embedding table.

### `CalculateTotalPrice`
Request shape:

```json
{
  "order_items": [
    { "product_id": 1, "quantity": 2 },
    { "product_id": 5, "quantity": 1 }
  ]
}
```

Response shape:

```json
{
  "total_price": 89.97
}
```

## Authentication

The service uses a gRPC auth interceptor with JWT bearer tokens.

### Public RPCs
These methods do not require authentication:

- `GetProductByID`
- `GetProductsByIDs`

### Protected RPCs
These methods require:

- gRPC metadata header `Authorization`
- value format: `Bearer <jwt>`

Write operations additionally expect the JWT claims to include a seller role.

## Kafka stock updates

At startup, the service creates a Kafka consumer for topic:

- `orders.placed`

Each consumed message is expected to contain ordered items. For each item, the service decreases product stock in PostgreSQL and refreshes the cached product entry.

## Environment variables

The service reads configuration from environment variables in `internal/config/config.go`.

| Variable | Required | Description |
|---|---:|---|
| `DSN` | Yes | PostgreSQL DSN |
| `BROKERS` | Yes | Comma-separated Kafka broker list |
| `CACHE_ADDR` | Yes | Redis address, e.g. `localhost:6379` |
| `CACHE_PASSWD` | No | Redis password |
| `PORT` | Yes | gRPC listen address, e.g. `:42069` |
| `EMBEDDING_COLLECTION_TABLE_NAME` | Yes | Embedding collection table name |
| `EMBEDDING_TABLE_NAME` | Yes | Embedding table name |
| `GEMINI_API_KEY` | Yes | Google Gemini API key for embeddings |
| `SECRET_KEY` | Yes | JWT signing secret |

## Prerequisites

Before running the service, make sure you have:

- Go `1.25+`
- PostgreSQL running and reachable through `DSN`
- Redis running
- Kafka running
- A valid Gemini API key
- The embedding tables already created in PostgreSQL

## Running locally

### 1. Export environment variables

Example:

```sh
export DSN='host=localhost port=5431 user=postgres password=postgres dbname=postgres sslmode=disable'
export CACHE_ADDR='localhost:6379'
export CACHE_PASSWD=''
export PORT=':42069'
export BROKERS='localhost:9093'
export SECRET_KEY='replace-with-a-secure-secret'
export GEMINI_API_KEY='replace-with-your-gemini-api-key'
export EMBEDDING_TABLE_NAME='langchain_pg_embedding'
export EMBEDDING_COLLECTION_TABLE_NAME='langchain_pg_collection'
```

### 2. Start the service

```sh
go run main.go
```

The server starts on the address configured by `PORT`.

## Running with Docker

Build the image:

```sh
docker build -t products-service .
```

Run the container:

```sh
docker run --rm -p 42069:42069 \
  -e DSN='host=host.docker.internal port=5431 user=postgres password=postgres dbname=postgres sslmode=disable' \
  -e CACHE_ADDR='host.docker.internal:6379' \
  -e CACHE_PASSWD='' \
  -e PORT=':42069' \
  -e BROKERS='host.docker.internal:9093' \
  -e SECRET_KEY='replace-with-a-secure-secret' \
  -e GEMINI_API_KEY='replace-with-your-gemini-api-key' \
  -e EMBEDDING_TABLE_NAME='langchain_pg_embedding' \
  -e EMBEDDING_COLLECTION_TABLE_NAME='langchain_pg_collection' \
  products-service
```

## Example `grpcurl` usage

### Get a product by ID

```sh
grpcurl -plaintext \
  -rpc-header 'product_id: 1' \
  localhost:42069 \
  products.ProductsService/GetProductByID
```

### Stream products by IDs

```sh
grpcurl -plaintext \
  -d '{"product_ids":[1,2,3]}' \
  localhost:42069 \
  products.ProductsService/GetProductsByIDs
```

### Add a product

```sh
grpcurl -plaintext \
  -rpc-header 'Authorization: Bearer <jwt>' \
  -d '{
    "name":"Wireless Mouse",
    "price":29.99,
    "image":"https://example.com/mouse.png",
    "category":"Electronics",
    "description":"Compact wireless mouse",
    "stock":20,
    "in_stock":true,
    "tags":["wireless","mouse","usb"]
  }' \
  localhost:42069 \
  products.ProductsService/AddProduct
```

### Calculate total price

```sh
grpcurl -plaintext \
  -rpc-header 'Authorization: Bearer <jwt>' \
  -d '{
    "order_items":[
      {"product_id":1,"quantity":2},
      {"product_id":5,"quantity":1}
    ]
  }' \
  localhost:42069 \
  products.ProductsService/CalculateTotalPrice
```

### Update a product

```sh
grpcurl -plaintext \
  -rpc-header 'Authorization: Bearer <jwt>' \
  -d '{
    "id":1,
    "name":"Wireless Mouse Pro",
    "price":39.99,
    "original_price":49.99,
    "image":"https://example.com/mouse-pro.png",
    "category":"Electronics",
    "description":"Updated product description",
    "stock":10,
    "in_stock":true,
    "tags":["wireless","mouse","pro"]
  }' \
  localhost:42069 \
  products.ProductsService/UpdateProduct
```

### Delete a product

```sh
grpcurl -plaintext \
  -rpc-header 'Authorization: Bearer <jwt>' \
  -d '{"id":1}' \
  localhost:42069 \
  products.ProductsService/DeleteProduct
```
