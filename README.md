# Order Service

A RESTful microservice for order management, part of the Release Orchestrator platform.

## Tech Stack

- **Language:** Go 1.25
- **Framework:** Gin
- **Database:** PostgreSQL (via pgx)
- **Config:** Environment variables + .env

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| POST | /api/v1/orders | Create order |
| GET | /api/v1/orders | List orders |
| GET | /api/v1/orders/:id | Get order |
| DELETE | /api/v1/orders/:id | Cancel order |

### Query Parameters for List

| Param | Type | Description |
|-------|------|-------------|
| user_id | UUID | Filter by user |

## Dependencies

- **User Service** — validates user existence
- **Payment Service** — processes payment for orders

## Quick Start

### Prerequisites

- Go 1.25+
- Docker (for PostgreSQL)

### Run locally

```bash
# Start PostgreSQL
docker compose up -d

# Run the service
make run
```

### Build Docker image

```bash
make docker-build VERSION=0.1.0
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | HTTP port |
| DB_HOST | localhost | PostgreSQL host |
| DB_PORT | 5432 | PostgreSQL port |
| DB_USER | postgres | Database user |
| DB_PASS | postgres | Database password |
| DB_NAME | order_db | Database name |
| USER_SERVICE_URL | http://localhost:8081 | User service URL |
| PAYMENT_SERVICE_URL | http://localhost:8082 | Payment service URL |

## Project Structure

```
.
├── cmd/main.go              # Entry point
├── internal/
│   ├── client/              # HTTP clients for external services
│   ├── config/              # Configuration
│   ├── handler/             # HTTP handlers
│   ├── model/               # Data models
│   ├── repository/          # Database layer
│   └── service/             # Business logic
├── migrations/              # SQL migrations
├── helm/                    # Kubernetes deployment
└── Dockerfile               # Multi-stage build
```
