HELP=This is a Task Manager microservice project

.PHONY: help
help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: install
install: ## Install dependencies
	go mod download
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: swagger
swagger: ## Generate swagger documentation
	swag init -g cmd/api/main.go -o docs

.PHONY: build
build: swagger ## Build the application
	go build -o bin/taskmanager ./cmd/api

.PHONY: run
run: swagger ## Run the application locally
	go run ./cmd/api/main.go

.PHONY: test
test: ## Run tests
	go test -v -race -coverprofile=coverage.txt -covermode=atomic $(shell go list ./... | grep -v -E '(cmd/api|cmd/loadtest|docs)')

.PHONY: test-coverage
test-coverage: test ## Run tests and show coverage
	go tool cover -func=coverage.txt
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: integration-test
integration-test: ## Run integration tests with Docker
	chmod +x scripts/test.sh
	./scripts/test.sh

.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t taskmanager:latest .

.PHONY: docker-up
docker-up: ## Start all services with docker-compose
	docker compose up -d

.PHONY: docker-down
docker-down: ## Stop all services
	docker compose down

.PHONY: docker-logs
docker-logs: ## View logs from all services
	docker compose logs -f

.PHONY: docker-restart
docker-restart: docker down docker up ## Restart all services

.PHONY: docker-rebuild
docker-rebuild: ## Stop, remove, rebuild and start all services
	docker compose down -v
	docker compose build --no-cache
	docker compose up -d

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf docs/
	rm -f coverage.txt coverage.html
	go clean

.PHONY: lint
lint: ## Run linter
	golangci-lint run

.PHONY: fmt
fmt: ## Format code
	go fmt ./...
	goimports -w .

.PHONY: benchmark
benchmark: ## Run benchmarks
	go test -bench=. -benchmem ./...

.PHONY: pprof
pprof: ## Run pprof CPU profiling
	go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./internal/repository
	@echo "CPU profile: cpu.prof"
	@echo "Memory profile: mem.prof"
	@echo "View with: go tool pprof cpu.prof"
