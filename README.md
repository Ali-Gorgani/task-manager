# Task Manager Microservice 📋

A production-ready RESTful API microservice for managing tasks (to-do items) built with Go, Gin framework, PostgreSQL, and Redis cache.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## 🎯 Features

### ✅ Core Features (Mandatory)
- **RESTful CRUD API** - Full Create, Read, Update, Delete operations for tasks
- **Gin Framework** - High-performance HTTP framework
- **PostgreSQL Database** - Reliable data persistence with proper indexing
- **TDD Approach** - >70% test coverage with unit and integration tests
- **Docker Support** - Multi-stage Dockerfile and Docker Compose setup
- **OpenAPI/Swagger** - Interactive API documentation
- **Observability** - Prometheus metrics with request tracking

### ⭐ Optional Features (Bonus)
- **Redis Cache** - Cache-aside pattern with automatic invalidation
- **Pagination & Filtering** - Filter by status and assignee with pagination
- **Load Testing** - Built-in load test tool and benchmark scenarios
- **pprof Integration** - CPU and memory profiling support

## 📁 Project Structure

```
task-manager/
├── cmd/
│   ├── api/           # Main application entry point
│   └── loadtest/      # Load testing tool
├── internal/
│   ├── cache/         # Redis cache implementation
│   ├── config/        # Configuration management
│   ├── handlers/      # HTTP handlers (controllers)
│   ├── metrics/       # Prometheus metrics
│   ├── models/        # Data models and DTOs
│   ├── repository/    # Database layer with interface
│   └── service/       # Business logic layer
├── scripts/           # Utility scripts
├── docs/              # Generated Swagger documentation
├── Dockerfile         # Multi-stage Docker build
├── docker-compose.yml # Complete stack setup
├── Makefile           # Build and development commands
└── README.md          # This file
```

## 🏗️ Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTP/REST
       ▼
┌─────────────────────────────────┐
│     Gin Router + Middleware     │
│  (Metrics, Logging, Recovery)   │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│         Handlers Layer          │
│    (HTTP Request/Response)      │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│        Service Layer            │
│     (Business Logic)            │
└────┬────────────────────────┬───┘
     │                        │
     ▼                        ▼
┌──────────┐           ┌──────────┐
│  Redis   │           │PostgreSQL│
│  Cache   │           │Repository│
└──────────┘           └──────────┘
```

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose (recommended)
- Go 1.25+ (for local development)
- PostgreSQL 15+ (if running locally)
- Redis 7+ (optional, for caching)

### 1. Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone <repository-url>
cd task-manager

# Start all services (API, PostgreSQL, Redis, Prometheus)
docker-compose up -d

# Check if services are running
docker-compose ps

# View logs
docker-compose logs -f api
```

The API will be available at `http://localhost:3000`

### 2. Local Development

```bash
# Install dependencies
make install

# Copy environment file for local development
cp .env.local .env

# Start PostgreSQL and Redis (via Docker)
docker-compose up -d postgres redis

# Generate Swagger docs
make swagger

# Run the application
make run
```

## 📚 API Documentation

### Interactive Documentation
Once the service is running, visit:
- **Swagger UI**: http://localhost:3000/swagger/index.html
- **OpenAPI JSON**: http://localhost:3000/swagger/doc.json

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check endpoint |
| GET | `/metrics` | Prometheus metrics |
| POST | `/api/v1/tasks` | Create a new task |
| GET | `/api/v1/tasks` | List all tasks (with filtering & pagination) |
| GET | `/api/v1/tasks/:id` | Get a specific task |
| PUT | `/api/v1/tasks/:id` | Update a task |
| DELETE | `/api/v1/tasks/:id` | Delete a task |

## 💡 Usage Examples

### Create a Task
```bash
curl -X POST http://localhost:3000/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Complete project documentation",
    "description": "Write comprehensive README and API docs",
    "status": "pending",
    "assignee": "john.doe@example.com"
  }'
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Complete project documentation",
  "description": "Write comprehensive README and API docs",
  "status": "pending",
  "assignee": "john.doe@example.com",
  "created_at": "2025-11-01T10:00:00Z",
  "updated_at": "2025-11-01T10:00:00Z"
}
```

### Get All Tasks (with Pagination)
```bash
curl "http://localhost:3000/api/v1/tasks?page=1&page_size=10"
```

**Response:**
```json
{
  "tasks": [...],
  "total": 100,
  "page": 1,
  "page_size": 10,
  "total_pages": 10
}
```

### Filter Tasks by Status
```bash
curl "http://localhost:3000/api/v1/tasks?status=pending"
```

### Filter by Assignee
```bash
curl "http://localhost:3000/api/v1/tasks?assignee=john.doe@example.com"
```

### Get a Specific Task
```bash
curl http://localhost:3000/api/v1/tasks/550e8400-e29b-41d4-a716-446655440000
```

### Update a Task
```bash
curl -X PUT http://localhost:3000/api/v1/tasks/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "status": "completed"
  }'
```

### Delete a Task
```bash
curl -X DELETE http://localhost:3000/api/v1/tasks/550e8400-e29b-41d4-a716-446655440000
```

## 🧪 Testing

### Run All Tests
```bash
make test
```

### Test Coverage Report
```bash
make test-coverage
```

This will generate:
- `coverage.txt` - Coverage data
- `coverage.html` - HTML coverage report

### Integration Tests
```bash
make integration-test
```

### Run Benchmarks
```bash
make benchmark
```

### CPU & Memory Profiling
```bash
make pprof
```

View profiles:
```bash
go tool pprof cpu.prof
go tool pprof mem.prof
```

## 📊 Monitoring & Observability

### Prometheus Metrics
Access metrics at: http://localhost:3000/metrics

Available metrics:
- `requests_total` - Total number of HTTP requests (by method, endpoint, status)
- `request_latency_histogram` - Request latency distribution
- `tasks_count` - Current number of tasks in the system

### Prometheus Dashboard
Access Prometheus at: http://localhost:9090

Example queries:
```promql
# Request rate
rate(requests_total[5m])

# Average latency
rate(request_latency_histogram_sum[5m]) / rate(request_latency_histogram_count[5m])

# Total tasks
tasks_count
```

## 🔥 Load Testing

Run the built-in load test:

```bash
# Make sure the service is running first
docker-compose up -d

# Run load test
go run cmd/loadtest/main.go
```

**Load Test Configuration:**
- Workers: 50 concurrent workers
- Total Requests: 1000 per test scenario
- Scenarios: Create, List, Filter tasks

**Example Output:**
```
Test 1: Create Tasks
Total Requests:       1000
Successful Requests:  1000 (100.00%)
Failed Requests:      0 (0.00%)
Total Duration:       2.5s
Avg Response Time:    125ms
Min Response Time:    45ms
Max Response Time:    340ms
Requests/Second:      400.00
```

## 🛠️ Development

### Available Make Commands

```bash
make help              # Show all available commands
make install           # Install dependencies
make swagger           # Generate Swagger docs
make build             # Build the application
make run               # Run locally
make test              # Run tests
make test-coverage     # Run tests with coverage report
make docker-build      # Build Docker image
make docker-up         # Start all services
make docker-down       # Stop all services
make docker-restart    # Restart all services
make docker-rebuild    # Stop, remove, rebuild and start all services (clean rebuild)
make docker-logs       # View service logs
make clean             # Clean build artifacts
make fmt               # Format code
make lint              # Run linter
```

### Configuration

The application uses **Viper** for flexible configuration management. Configuration files:

- **`.env.local`** - Template for local development (uses `localhost`)
- **`.env.container`** - Template for containerized environments (uses Docker service names: `postgres`, `redis`)
- **`.env`** - Active config file (gitignored) - copy from templates based on your environment

**For Local Development:**
```bash
# Copy the local template to .env
cp .env.local .env

# The .env file will have localhost URLs:
# DATABASE_URL=postgres://postgres:postgres@localhost:5432/taskmanager?sslmode=disable
# REDIS_URL=localhost:6379

# Run the application
make run
```

**For Docker Compose:**
```bash
# Copy the container template to .env
cp .env.container .env

# The .env file will have Docker service names:
# DATABASE_URL=postgres://postgres:postgres@postgres:5432/taskmanager?sslmode=disable
# REDIS_URL=redis:6379

# Start all services
docker-compose up -d
```

**For Production (Environment Variables):**
```bash
export SERVER_PORT=3000
export DATABASE_URL="postgres://user:pass@prod-host:5432/taskmanager?sslmode=disable"
export REDIS_URL="prod-redis:6379"
export REDIS_PASSWORD="secure-password"
export ENVIRONMENT="production"
```

**Configuration Priority:** Environment variables > `.env` file > Default values

**Note:** `.env` is gitignored for security. Always copy from examples.

## 📈 Performance

### Benchmarks
Run benchmarks with:
```bash
go test -bench=. -benchmem ./...
```

### Database Indexes
The following indexes are created for optimal performance:
- `idx_tasks_status` - Status filtering
- `idx_tasks_assignee` - Assignee filtering
- `idx_tasks_created_at` - Sorting by creation date

## 🎯 Design Decisions & Trade-offs

### ✅ Architectural Decisions

1. **Repository Pattern**
   - ✅ Easy to mock for testing
   - ✅ Separation of concerns
   - ✅ Can swap database implementations
   - ⚠️ Slight overhead from abstraction

2. **Cache-Aside Pattern with Redis**
   - ✅ Reduces database load
   - ✅ Improved read performance
   - ⚠️ Cache invalidation complexity
   - ⚠️ Eventual consistency

3. **UUID for Task IDs**
   - ✅ Globally unique
   - ✅ No collision risk
   - ✅ Can be generated client-side
   - ⚠️ Larger than auto-increment IDs
   - ⚠️ Not human-readable

4. **Multi-stage Docker Build**
   - ✅ Small final image (~15MB)
   - ✅ Fast deployment
   - ✅ Security (no build tools in runtime)
   - ⚠️ Longer build time

### ⚖️ Trade-offs

| Aspect | Decision | Pros | Cons |
|--------|----------|------|------|
| Database | PostgreSQL | ACID compliant, mature | Slightly heavier than MySQL |
| Cache | Redis | Fast, feature-rich | Additional infrastructure |
| ID Strategy | UUID v4 | Distributed-friendly | Not sequential |
| HTTP Framework | Gin | Fast, minimal | Less batteries-included than Echo |

### 🔮 Future Improvements
- [ ] Add authentication & authorization (JWT)
- [ ] Implement task priorities and due dates
- [ ] Add task assignment notifications
- [ ] Implement task comments/history
- [ ] Add GraphQL API support
- [ ] Kubernetes deployment manifests
- [ ] CI/CD pipeline configuration
- [ ] Distributed tracing with Jaeger

## 🐛 Troubleshooting

### Service won't start
```bash
# Check if ports are already in use
lsof -i :3000
lsof -i :5432
lsof -i :6379

# View service logs
docker-compose logs api
```

### Database connection errors
```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Connect to PostgreSQL directly
docker-compose exec postgres psql -U postgres -d taskmanager
```

### Redis connection failures
The service will work without Redis (caching disabled). Check Redis with:
```bash
docker-compose exec redis redis-cli ping
```

## 📄 License

This project is licensed under the MIT License.

## 👤 Author

Ali Gorgani

---
