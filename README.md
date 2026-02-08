# Retail Management System (RMS)

A high-performance, local-first retail inventory management system built with Go, featuring tamper-proof audit logging and flexible schema design.

## Architecture

This system follows **Clean/Hexagonal Architecture** principles with strict separation of concerns:

```
├── cmd/server              # Application entry point
├── internal/
│   ├── core/
│   │   ├── domain/         # Pure business entities (Product, AuditLog)
│   │   ├── ports/          # Interface definitions (Repository, Service)
│   │   └── services/       # Business logic implementations
│   └── adapters/
│       ├── handler/        # HTTP handlers (REST API)
│       └── storage/        # SQLite repository implementations
└── migrations/             # Database schema
```

## Features

### 1. Flexible Inventory Schema
- **Entity-Attribute-Value (EAV) Hybrid**: Products have a `properties` JSON column for category-specific attributes
- Store electrical components, liquor, pharmaceuticals, or any product type without schema migrations
- Example properties: `{"voltage": "220v", "amperage": "5A"}` for electrical components

### 2. Tamper-Proof Audit Logging
- **Blockchain-like Hash Chain**: Each audit entry's hash depends on the previous entry
- Formula: `current_hash = SHA256(payload + timestamp + prev_hash)`
- Chain verification endpoint to detect tampering
- Any modification or deletion breaks the chain and is detectable

### 3. Clean Architecture Benefits
- **Framework Independence**: Business logic is isolated from frameworks
- **Testability**: Pure domain entities with no external dependencies
- **Database Agnostic**: Easy to switch from SQLite to PostgreSQL
- **Interface-Driven**: All layers communicate through well-defined interfaces

## Tech Stack

- **Language**: Go 1.21+
- **Database**: SQLite (local-first, replication-ready)
- **HTTP Framework**: Fiber v2
- **Database Access**: sqlx (raw SQL, no ORM)
- **SQLite Driver**: go-sqlite3

## Quick Start

### Prerequisites
- Go 1.21 or higher
- SQLite (included via CGO)

### Installation

```bash
# Clone the repository
git clone https://github.com/torantous1337/retail-management.git
cd retail-management

# Download dependencies
go mod tidy

# Build the server
go build -o bin/server ./cmd/server

# Run the server
./bin/server
```

The server will start on port 8080 by default.

### Environment Variables

- `PORT`: Server port (default: 8080)
- `DB_PATH`: SQLite database file path (default: retail.db)

Example:
```bash
PORT=3000 DB_PATH=/data/retail.db ./bin/server
```

## API Endpoints

### Health Check
```bash
GET /health
```

### Products

#### Create Product
```bash
POST /api/v1/products
Content-Type: application/json

{
  "name": "Industrial Circuit Breaker",
  "sku": "CB-220-50A",
  "base_price": 45.99,
  "properties": {
    "voltage": "220v",
    "amperage": "50A",
    "poles": 3,
    "manufacturer": "ABB"
  }
}
```

#### List Products
```bash
GET /api/v1/products?limit=10&offset=0
```

#### Get Product by ID
```bash
GET /api/v1/products/{id}
```

#### Get Product by SKU
```bash
GET /api/v1/products/sku/{sku}
```

#### Update Product
```bash
PUT /api/v1/products/{id}
Content-Type: application/json

{
  "name": "Updated Name",
  "sku": "CB-220-50A",
  "base_price": 49.99,
  "properties": {...}
}
```

#### Delete Product
```bash
DELETE /api/v1/products/{id}
```

### Audit Logs

#### List Audit Logs
```bash
GET /api/v1/audit-logs?limit=10&offset=0
```

#### Verify Audit Chain
```bash
GET /api/v1/audit-logs/verify

Response:
{
  "valid": true
}
```

## Database Schema

### Products Table
```sql
CREATE TABLE products (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    sku TEXT UNIQUE NOT NULL,
    base_price REAL NOT NULL DEFAULT 0.0,
    properties TEXT,  -- JSON column for flexible attributes
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### Audit Logs Table
```sql
CREATE TABLE audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    action TEXT NOT NULL,
    user_id TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    payload TEXT NOT NULL,  -- JSON payload
    prev_hash TEXT,  -- Hash of previous record
    current_hash TEXT NOT NULL  -- SHA256(payload + timestamp + prev_hash)
);
```

## Security Features

### Audit Trail Verification
The system maintains a cryptographic audit trail that makes tampering evident:

1. Each audit entry contains a hash of: payload + timestamp + previous_hash
2. To verify the chain, call `/api/v1/audit-logs/verify`
3. Any modification to a log entry breaks the chain
4. Deleting entries leaves gaps in the chain

This creates an immutable audit trail similar to blockchain technology.

## Testing the System

```bash
# Start the server
./bin/server

# Create a product (in another terminal)
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "LED Bulb 10W",
    "sku": "LED-10W-WW",
    "base_price": 5.99,
    "properties": {
      "wattage": "10W",
      "color_temperature": "warm_white",
      "lumens": 800
    }
  }'

# List products
curl http://localhost:8080/api/v1/products

# Verify audit chain
curl http://localhost:8080/api/v1/audit-logs/verify
```

## Development

### Project Structure Philosophy

- **Domain Layer** (`internal/core/domain`): Pure business entities with zero dependencies
- **Ports Layer** (`internal/core/ports`): Interface definitions for repositories and services
- **Services Layer** (`internal/core/services`): Business logic implementations
- **Adapters Layer** (`internal/adapters`): External world integrations (HTTP, Database)
- **Main** (`cmd/server`): Dependency injection and application wiring

### Design Decisions

1. **No ORM**: Raw SQL provides full control over queries for performance tuning
2. **Interface-Based**: All dependencies are injected via interfaces for testability
3. **JSON Properties**: Flexible schema without migrations, stored as TEXT in SQLite
4. **SHA256 Hashing**: Industry-standard cryptographic hashing for audit trails
5. **Fiber Framework**: High-performance HTTP framework with Express-like API

## Future Enhancements

- gRPC endpoints for internal service communication
- Litestream integration for real-time SQLite replication
- PostgreSQL adapter for production deployments
- Authentication and authorization middleware
- GraphQL API layer
- Prometheus metrics and observability
- Docker containerization

## License

MIT License - see LICENSE file for details