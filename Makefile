.PHONY: all build test lint clean migrate-up migrate-down run dev help fmt

# Variables
BINARY_NAME=conduit
GO=go
GOLANGCI_LINT=golangci-lint
MIGRATE=migrate
MIGRATION_DIR=migrations
AIR=air
GOFUMPT=gofumpt

# Default target
all: build

# Build the application
build:
	@echo "Building application..."
	$(GO) build -o $(BINARY_NAME) ./cmd/...

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage
test-coverage: test
	$(GO) tool cover -html=coverage.out

# Format code
fmt:
	@echo "Formatting code with gofumpt..."
	$(GOFUMPT) -l -w .

# Run linter
lint:
	@echo "Running linter..."
	$(GOLANGCI_LINT) run

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out

# Database migrations
migrate-up:
	@echo "Running database migrations up..."
	$(MIGRATE) -path $(MIGRATION_DIR) -database "postgres://postgres:postgres@localhost:5432/conduit?sslmode=disable" up

migrate-down:
	@echo "Running database migrations down..."
	$(MIGRATE) -path $(MIGRATION_DIR) -database "postgres://postgres:postgres@localhost:5432/conduit?sslmode=disable" down

# Run the application
run:
	@echo "Running application..."
	$(GO) run ./cmd/...

# Run the application with hot reload (using Air)
dev:
	@echo "Starting development server with hot reload..."
	$(AIR)

# Help target
help:
	@echo "Available targets:"
	@echo "  all           - Build the application (default)"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  fmt           - Format code with gofumpt"
	@echo "  lint          - Run linter"
	@echo "  clean         - Remove build artifacts"
	@echo "  migrate-up    - Run database migrations up"
	@echo "  migrate-down  - Run database migrations down"
	@echo "  run           - Run the application"
	@echo "  dev           - Run the application with hot reload (using Air)"
	@echo "  help          - Show this help message"
