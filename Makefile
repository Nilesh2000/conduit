.PHONY: all build cleanrun dev test test-coverage fmt lint migrate-up migrate-down help

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

# Build application
build:
	@echo "Building application..."
	$(GO) build -o $(BINARY_NAME) ./cmd/...

clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME) coverage.out

# Run application
run:
	@echo "Running application..."
	$(GO) run ./cmd/...

# Run with hot reload (Air)
dev:
	@echo "Starting development server with hot reload..."
	$(AIR)

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

# View test coverage in browser
test-coverage: test
	$(GO) tool cover -html=coverage.out

# Format code
fmt:
	@echo "Formatting code with gofumpt and golines..."
	$(GOFUMPT) -l -w .
	$(GOLINES) -w .

# Run linter with gosec
lint:
	@echo "Running linter..."
	$(GOLANGCI_LINT) run --enable=gosec

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
	@echo "  all           - Build application (default)"
	@echo "  build         - Build application"
	@echo "  run           - Run application"
	@echo "  dev           - Run with hot reload (Air)"
	@echo "  clean         - Remove build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-coverage - View test coverage in browser"
	@echo "  fmt           - Format code (gofumpt + golines)"
	@echo "  lint          - Run linter with gosec"
	@echo "  migrate-up    - Run database migrations up"
	@echo "  migrate-down  - Run database migrations down"
	@echo "  help          - Show this help message"
