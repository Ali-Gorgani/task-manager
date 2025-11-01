# Task Manager - Quick Start Guide

## Prerequisites
- Docker & Docker Compose installed
- Git

## Quick Start (3 minutes ⏱️)

### 1. Clone & Start
```bash
# Clone the repository
git clone <repository-url>
cd task-manager

# Start all services
docker-compose up -d

# Wait for services to be ready (about 30 seconds)
docker-compose logs -f api
```

### 2. Test the API
```bash
# Health check
curl http://localhost:3000/health

# Create a task
curl -X POST http://localhost:3000/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"My First Task","status":"pending","assignee":"me@example.com"}'

# List all tasks
curl http://localhost:3000/api/v1/tasks

# View API Documentation
open http://localhost:3000/swagger/index.html
```

### 3. View Metrics
```bash
# Prometheus metrics
curl http://localhost:3000/metrics

# Prometheus UI
open http://localhost:9090
```

## Running Tests

```bash
# Run all tests with coverage
make test-coverage

# Run integration tests
make integration-test

# Run benchmarks
make benchmark
```

## Load Testing

```bash
# Run load test
go run cmd/loadtest/main.go
```

## Stop Services

```bash
docker-compose down
```

## Development Mode

```bash
# Start database only
docker-compose up -d postgres redis

# Run API locally
make run
```

## Troubleshooting

### Port already in use
```bash
# Find process using port 3000
lsof -i :3000
# Kill it or change SERVER_PORT in .env
```

### Database connection failed
```bash
# Check PostgreSQL is running
docker-compose ps postgres
docker-compose logs postgres
```

### Redis connection failed
```bash
# Service will work without Redis (no caching)
docker-compose ps redis
```

## Project Structure Summary

```
task-manager/
├── cmd/api/           # Main application
├── internal/
│   ├── handlers/      # HTTP handlers
│   ├── service/       # Business logic
│   ├── repository/    # Database access
│   ├── cache/         # Redis caching
│   ├── models/        # Data models
│   ├── config/        # Configuration
│   └── metrics/       # Prometheus metrics
├── docs/              # Swagger documentation
├── scripts/           # Utility scripts
├── Dockerfile         # Multi-stage build
├── docker-compose.yml # Services orchestration
├── Makefile           # Build commands
└── README.md          # Full documentation
```

## Key Features Implemented

✅ **Mandatory Requirements**
- RESTful CRUD API with Gin framework
- PostgreSQL database with migrations
- Unit & Integration tests (50%+ coverage)
- Docker & Docker Compose setup
- OpenAPI/Swagger documentation
- Prometheus metrics & observability

⭐ **Bonus Features**
- Redis cache with cache-aside pattern
- Pagination & filtering (status, assignee)
- Load testing tool
- Benchmark tests
- pprof profiling support

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/tasks` | Create task |
| GET | `/api/v1/tasks` | List tasks (paginated, filterable) |
| GET | `/api/v1/tasks/:id` | Get task by ID |
| PUT | `/api/v1/tasks/:id` | Update task |
| DELETE | `/api/v1/tasks/:id` | Delete task |
| GET | `/health` | Health check |
| GET | `/metrics` | Prometheus metrics |
| GET | `/swagger/*` | API documentation |

## Test Coverage

Run tests with coverage:
```bash
make test-coverage
```

Current coverage: ~50% (focused on core business logic)
- models: 100%
- config: 100%
- metrics: 90%
- repository: 75%
- service: 62%

## Performance

Load test results (50 workers, 1000 requests):
- Avg Response Time: ~125ms
- Throughput: ~400 req/s
- Success Rate: 100%

## Support

For issues or questions, see the full [README.md](README.md) or [ARCHITECTURE.md](ARCHITECTURE.md)
