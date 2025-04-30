# Variables
BIN := "conduit"
GO := "go"
GOLANGCI_LINT := "golangci-lint"
MIGRATE := "migrate"
MIGRATION_DIR := "migrations"
AIR := "air"
GOFUMPT := "gofumpt"
GOLINES := "golines"
DB_URL := "postgres://postgres:postgres@localhost:5432/conduit?sslmode=disable"

# Default recipe
default:
    @just --list

# Setup development environment
setup:
    #!/usr/bin/env bash
    echo "Installing development tools..."
    {{GO}} install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    {{GO}} install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    {{GO}} install github.com/air-verse/air@latest
    {{GO}} install mvdan.cc/gofumpt@latest
    {{GO}} install github.com/segmentio/golines@latest
    {{GO}} install github.com/evilmartians/lefthook@latest
    echo "Setting up git hooks..."
    lefthook install
    echo "Installing dependencies..."
    just deps
    echo "Setup complete!"

# Initialize project
init:
    #!/usr/bin/env bash
    echo "Initializing project..."
    echo "Creating .env file..."
    test -f .env || cp .env.example .env
    echo "Running database migrations..."
    {{MIGRATE}} -path {{MIGRATION_DIR}} -database "{{DB_URL}}" up
    echo "Project initialization complete!"

# Manage dependencies
deps:
    #!/usr/bin/env bash
    echo "Installing dependencies..."
    {{GO}} mod download
    {{GO}} mod tidy

# Build application
build: fmt lint
    #!/usr/bin/env bash
    echo "Building application..."
    {{GO}} build -o {{BIN}} ./cmd/...

# Clean build artifacts and test files
clean:
    #!/usr/bin/env bash
    echo "Cleaning..."
    rm -f {{BIN}}
    rm -f coverage.out

# Run application
run:
    #!/usr/bin/env bash
    echo "Running application..."
    {{GO}} run ./cmd/...

# Run with hot reload (Air)
dev:
    #!/usr/bin/env bash
    echo "Starting development server with hot reload..."
    {{AIR}}

# Run with Docker
docker:
    #!/usr/bin/env bash
    echo "Starting Docker containers..."
    docker-compose up -d

# Run tests
test:
    #!/usr/bin/env bash
    echo "Running tests..."
    {{GO}} test -p 4 -v -race -coverprofile=coverage.out ./...

# View test coverage report in browser
test-coverage: test
    #!/usr/bin/env bash
    {{GO}} tool cover -html=coverage.out

# Format code
fmt:
    #!/usr/bin/env bash
    echo "Formatting code with gofumpt and golines..."
    {{GOFUMPT}} -l -w .
    {{GOLINES}} -w .

# Run linter with gosec security checks
lint:
    #!/usr/bin/env bash
    echo "Running linter..."
    {{GOLANGCI_LINT}} run --enable=gosec ./...

# Create new migration files
create-migration name:
    #!/usr/bin/env bash
    echo "Creating migration files..."
    {{MIGRATE}} create -ext sql -dir {{MIGRATION_DIR}} -seq {{name}}

# Run database migrations up
migrate-up:
    #!/usr/bin/env bash
    echo "Running database migrations up..."
    {{MIGRATE}} -path {{MIGRATION_DIR}} -database "{{DB_URL}}" up

# Run database migrations down
migrate-down:
    #!/usr/bin/env bash
    echo "Running database migrations down..."
    {{MIGRATE}} -path {{MIGRATION_DIR}} -database "{{DB_URL}}" down

# View Go documentation
godoc:
    #!/usr/bin/env bash
    echo "Starting godoc server at http://localhost:6060"
    godoc -http=:6060

# List all available recipes
list:
    @just --list
