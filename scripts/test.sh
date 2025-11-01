#!/bin/bash
# Integration test script that runs tests with a real database

set -e

echo "Starting integration tests..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running"
    exit 1
fi

# Cleanup function
cleanup() {
    echo "Cleaning up..."
    docker stop taskmanager-test-db 2>/dev/null || true
    docker rm taskmanager-test-db 2>/dev/null || true
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Clean up any existing test container
cleanup

# Start test database
echo "Starting PostgreSQL test database..."
docker run -d --name taskmanager-test-db \
    -e POSTGRES_USER=testuser \
    -e POSTGRES_PASSWORD=testpass \
    -e POSTGRES_DB=taskmanager_test \
    -p 5434:5432 \
    postgres:15-alpine

# Wait for database to be ready
echo "Waiting for database to be ready..."
for i in {1..30}; do
    if docker exec taskmanager-test-db pg_isready -U testuser > /dev/null 2>&1; then
        echo "Database is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "Error: Database failed to start"
        exit 1
    fi
    sleep 1
done

# Set test environment variables
export DATABASE_URL="postgres://testuser:testpass@localhost:5434/taskmanager_test?sslmode=disable"
export REDIS_URL="localhost:6379"
export ENVIRONMENT="test"

# Run tests (excluding cmd/api, cmd/loadtest, and docs from coverage)
echo "Running integration tests..."
go test -v -race -coverprofile=coverage.txt -covermode=atomic $(go list ./... | grep -v -E '(cmd/loadtest|docs)')

# Calculate coverage
COVERAGE=$(go tool cover -func=coverage.txt | grep total | awk '{print $3}')
echo ""
echo "=========================================="
echo "Total test coverage: $COVERAGE"
echo "=========================================="

# Generate HTML coverage report
go tool cover -html=coverage.txt -o coverage.html
echo "HTML coverage report generated: coverage.html"

echo ""
echo "Integration tests completed successfully!"
