.PHONY: help build test run clean docker-build docker-run docker-stop lint

# Default target
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  test         - Run all tests"
	@echo "  test-unit    - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  run          - Run the application locally"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  demo         - Setup DB, run server, and exercise API (no Docker)"
	@echo "  docker-stop  - Stop Docker Compose services"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"

# Build the application
build:
	@echo "Building application..."
	go build -o bin/server ./cmd/server

# Run all tests
test:
	@echo "Running all tests..."
	go test -v ./...

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	go test -v ./internal/delivery

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	go test -v ./internal/campaigns

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -cover ./...

# Run the application locally
run:
	@echo "Running application..."
	go run ./cmd/server/main.go

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker-compose build

# Run with Docker Compose
docker-run:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

# Stop Docker Compose services
docker-stop:
	@echo "Stopping Docker Compose services..."
	docker-compose down

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Database operations
db-migrate:
	@echo "Running database migrations..."
	docker exec -i $$(docker-compose ps -q postgres) psql -U postgres -d targeting_db < db/migrations/init.sql
	docker exec -i $$(docker-compose ps -q postgres) psql -U postgres -d targeting_db < db/migrations/seed.sql

db-reset:
	@echo "Resetting database..."
	docker-compose down -v
	docker-compose up postgres -d
	sleep 5
	$(MAKE) db-migrate

# Development setup
dev-setup: deps docker-run
	@echo "Waiting for database to be ready..."
	sleep 10
	$(MAKE) db-migrate
	@echo "Development environment ready!"

# Production build
prod-build:
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server ./cmd/server 

demo:
	@echo "Running local demo (no Docker)..."
	bash ./scripts/setup-and-demo.sh 