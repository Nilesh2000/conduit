.PHONY: all build cleanrun dev test test-coverage fmt vet lint migrate-up migrate-down help

# Variables
BINARY_NAME=conduit
GO=go
GOLANGCI_LINT=golangci-lint
MIGRATE=migrate
MIGRATION_DIR=migrations
AIR=air
GOFUMPT=gofumpt
GOLINES=golines
DB_URL=postgres://postgres:postgres@localhost:5432/conduit?sslmode=disable

# Default target
all: build

# Build the application
build:
	@echo "Building application..."
	$(GO) build -o $(BINARY_NAME) ./cmd/...

clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME) coverage.out

# Run targets
run:
	@echo "Running application..."
	$(GO) run ./cmd/...

dev:
	@echo "Starting development server with hot reload..."
	$(AIR)

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage
test-coverage: test
	$(GO) tool cover -html=coverage.out

# Format code
fmt:
	@echo "Formatting code with gofumpt and golines..."
	$(GOFUMPT) -l -w .
	$(GOLINES) -w .

# Run vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

# Run linter
lint:
	@echo "Running linter..."
	$(GOLANGCI_LINT) run

# Database migrations
migrate-up:
	@echo "Running database migrations up..."
	$(MIGRATE) -path $(MIGRATION_DIR) -database "$(DB_URL)" up

migrate-down:
	@echo "Running database migrations down..."
	$(MIGRATE) -path $(MIGRATION_DIR) -database "$(DB_URL)" down

# Help target
help:
	@echo "Available targets:"
	@echo "  all           - Build the application (default)"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  dev           - Run the application with hot reload (using Air)"
	@echo "  clean         - Remove build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  fmt           - Format code with gofumpt and golines"
	@echo "  vet           - Run go vet for static analysis"
	@echo "  lint          - Run linter"
	@echo "  migrate-up    - Run database migrations up"
	@echo "  migrate-down  - Run database migrations down"
	@echo "  help          - Show this help message"
