.PHONY: all setup init deps build clean run dev docker test test-coverage fmt lint create-migration migrate-up migrate-down help

# Variables
BIN=conduit
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

# Setup development environment
setup:
	@echo "Installing development tools..."
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	$(GO) install github.com/air-verse/air@latest
	$(GO) install mvdan.cc/gofumpt@latest
	$(GO) install github.com/segmentio/golines@latest
	$(GO) install github.com/evilmartians/lefthook@latest
	@echo "Setting up git hooks..."
	lefthook install
	@echo "Installing dependencies..."
	$(MAKE) deps
	@echo "Setup complete!"

# Initialize project
init:
	@echo "Initializing project..."
	@echo "Creating .env file..."
	@test -f .env || cp .env.example .env
	@echo "Running database migrations..."
	$(MIGRATE) -path $(MIGRATION_DIR) -database "$(DB_URL)" up
	@echo "Project initialization complete!"

# Manage dependencies
deps:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# Build application
build:
	@echo "Building application..."
	$(GO) build -o $(BIN) ./cmd/...

# Clean build artifacts and test files
clean:
	@echo "Cleaning..."
	rm -f $(BIN)
	rm -f coverage.out

# Run application
run:
	@echo "Running application..."
	$(GO) run ./cmd/...

# Run with hot reload (Air)
dev:
	@echo "Starting development server with hot reload..."
	$(AIR)

# Run with Docker
docker:
	@echo "Starting Docker containers..."
	docker-compose up -d

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

# View test coverage report in browser
test-coverage: test
	$(GO) tool cover -html=coverage.out

# Format code
fmt:
	@echo "Formatting code with gofumpt and golines..."
	$(GOFUMPT) -l -w .
	$(GOLINES) -w .

# Run linter with gosec security checks
lint:
	@echo "Running linter..."
	$(GOLANGCI_LINT) run --enable=gosec

# Create new migration files
create-migration:
	@if [ -z "$(name)" ]; then \
		echo "Error: Migration name is required. Usage: make create-migration name=migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration files..."
	$(MIGRATE) create -ext sql -dir $(MIGRATION_DIR) -seq $(name)

# Run database migrations up
migrate-up:
	@echo "Running database migrations up..."
	$(MIGRATE) -path $(MIGRATION_DIR) -database "$(DB_URL)" up

# Run database migrations down
migrate-down:
	@echo "Running database migrations down..."
	$(MIGRATE) -path $(MIGRATION_DIR) -database "$(DB_URL)" down

# Help target
help:
	@echo "Available targets:"
	@echo "  all              - Build application (default)"
	@echo "  setup            - Install development tools"
	@echo "  init             - Initialize project (env, db)"
	@echo "  deps             - Install and tidy dependencies"
	@echo "  build            - Build application"
	@echo "  clean            - Remove build artifacts"
	@echo "  run              - Run application"
	@echo "  dev              - Run with hot reload (Air)"
	@echo "  docker           - Run with Docker"
	@echo "  test             - Run tests"
	@echo "  test-coverage    - View test coverage in browser"
	@echo "  fmt              - Format code (gofumpt + golines)"
	@echo "  lint             - Run linter with gosec"
	@echo "  create-migration - Create new migration files (requires name parameter)"
	@echo "  migrate-up       - Run database migrations up"
	@echo "  migrate-down     - Run database migrations down"
	@echo "  help             - Show this help message"
