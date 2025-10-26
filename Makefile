# Developer Portal Backend Makefile

# Variables
APP_NAME=developer-portal-backend
DOCKER_COMPOSE_FILE=docker/docker-compose.yml
MIGRATION_DIR=internal/database/migrations

# Go related variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=main
BINARY_PATH=./$(BINARY_NAME)

# Docker related variables
DOCKER_BUILD_CONTEXT=.
DOCKERFILE_PATH=docker/Dockerfile

.PHONY: help build clean test deps run run-dev run-dev-quick dev docker-up docker-down docker-build migrate-up migrate-down migrate-create lint format \
        test-unit test-integration test-all test-coverage test-race test-bench test-setup test-teardown test-db-up test-db-down \
        test-docker test-docker-up test-docker-down test-clean test-verbose mocks

# Default target
help: ## Show this help message
	@echo "🏗️  Developer Portal Backend - Available Commands"
	@echo ""
	@echo "🚀 Quick Start:"
	@echo "   make setup          - Complete development environment setup"
	@echo "   make run-dev        - Start in development mode (resets DB & loads data)"
	@echo "   make run-dev-quick  - Start in development mode (keeps existing DB)"
	@echo "   make run            - Start in production mode (requires JWT_SECRET env var)"
	@echo ""
	@echo "📋 All Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "🛡️  Security Note:"
	@echo "   • Development: JWT secrets are auto-generated per session"
	@echo "   • Production: Set JWT_SECRET environment variable with secure value"
	@echo "   • Never commit secrets to version control!"

# Development commands
build: ## Build the application
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/server

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

test: ## Run tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

# JWT Secret Management
check-jwt-secret: ## Validate JWT secret configuration
	@if [ -z "$$JWT_SECRET" ]; then \
		echo "❌ JWT_SECRET not found"; \
		echo ""; \
		echo "🚨 MISSING JWT SECRET!"; \
		echo ""; \
		echo "📖 What is JWT_SECRET?"; \
		echo "   JWT_SECRET is a cryptographic signing key for authentication tokens."; \
		echo "   Without it, the application cannot:"; \
		echo "   • Generate secure user session tokens"; \
		echo "   • Validate incoming authentication requests"; \
		echo "   • Protect API endpoints from unauthorized access"; \
		echo ""; \
		echo "🔧 How to fix:"; \
		echo "   For DEVELOPMENT: make run-dev  # Auto-generates secure secret"; \
		echo "   For PRODUCTION:  export JWT_SECRET=\$$(openssl rand -base64 32)"; \
		echo ""; \
		echo "⚠️  SECURITY NOTE:"; \
		echo "   • Use different secrets for dev/staging/production"; \
		echo "   • Never commit secrets to version control"; \
		echo "   • Store production secrets in secure vaults"; \
		echo ""; \
		exit 1; \
	else \
		echo "✅ JWT secret found in environment variable"; \
	fi

run: check-jwt-secret build ## Production run (requires JWT_SECRET env var)
	@echo "🚀 Starting Developer Portal Backend (Production Mode)..."
	@echo "✅ All security checks passed"
	./$(BINARY_NAME)

run-dev: db-reset build load-initial-data ## Development run (uses JWT_SECRET from env or auto-generates; resets DB to avoid schema drift)
	@echo "🔐 Configuring JWT secret for development..."
	@if [ -z "$$JWT_SECRET" ] && [ -f .env ]; then \
		set -a; . ./.env; set +a; \
	fi; \
	if [ -z "$$JWT_SECRET" ]; then \
		JWT_SECRET=$$(openssl rand -base64 32); \
		echo "✅ Generated NEW JWT secret: $$JWT_SECRET"; \
		echo "💡 Tip: Set JWT_SECRET in .env to persist tokens across restarts"; \
	else \
		echo "✅ Using EXISTING JWT secret from .env file"; \
		echo "🔑 Tokens will persist across server restarts"; \
	fi; \
	echo "🚀 Starting Developer Portal Backend (Development Mode)..."; \
	echo "📋 Development Configuration:"; \
	echo "   • JWT Secret: ✅ Configured"; \
	echo "   • Database: Check docker-compose.yml for settings"; \
	echo "   • Auth Providers: Check config/auth.yaml"; \
	echo "   • Initial Data: ✅ Loaded initial teams"; \
	echo ""; \
	JWT_SECRET=$$JWT_SECRET ./$(BINARY_NAME)

run-dev-quick: db-up build ## Quick development run (uses JWT_SECRET from env or auto-generates; keeps existing DB)
	@echo "🔐 Configuring JWT secret for development..."
	@if [ -z "$$JWT_SECRET" ] && [ -f .env ]; then \
		set -a; . ./.env; set +a; \
	fi; \
	if [ -z "$$JWT_SECRET" ]; then \
		JWT_SECRET=$$(openssl rand -base64 32); \
		echo "✅ Generated NEW JWT secret: $$JWT_SECRET"; \
		echo "💡 Tip: Set JWT_SECRET in .env to persist tokens across restarts"; \
	else \
		echo "✅ Using EXISTING JWT secret from .env file"; \
		echo "🔑 Tokens will persist across server restarts"; \
	fi; \
	echo "🚀 Starting Developer Portal Backend (Development Mode - Quick)..."; \
	echo "📋 Development Configuration:"; \
	echo "   • JWT Secret: ✅ Configured"; \
	echo "   • Database: ✅ Using existing database"; \
	echo "   • Auth Providers: Check config/auth.yaml"; \
	echo ""; \
	JWT_SECRET=$$JWT_SECRET ./$(BINARY_NAME)

dev: ## Run the application in development mode with hot reload (requires air)
	@if command -v air > /dev/null; then \
		if [ -z "$$JWT_SECRET" ] && [ -f .env ]; then \
			set -a; . ./.env; set +a; \
		fi; \
		if [ -z "$$JWT_SECRET" ]; then \
			echo "❌ JWT_SECRET environment variable not set"; \
			echo "🔧 For development with hot reload:"; \
			echo "   Set JWT_SECRET in .env file, or"; \
			echo "   export JWT_SECRET=\$$(openssl rand -base64 32)"; \
			echo "   make dev"; \
			echo ""; \
			echo "🔧 Or use: make run-dev  # (no hot reload but auto-generates secret)"; \
			exit 1; \
		fi; \
		echo "✅ Using JWT secret from .env file"; \
		air; \
	else \
		echo "Air not found. Install it with: go install github.com/cosmtrek/air@latest"; \
		echo "Or run 'make run-dev' for development with auto-generated JWT secret"; \
	fi

# Docker commands
docker-build: ## Build Docker image
	docker build -f $(DOCKERFILE_PATH) -t $(APP_NAME):latest $(DOCKER_BUILD_CONTEXT)

docker-up: ## Start all services with Docker Compose
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d

docker-down: ## Stop all services
	docker-compose -f $(DOCKER_COMPOSE_FILE) down

docker-logs: ## View logs from all services
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

docker-restart: docker-down docker-up ## Restart all services

# Database commands
db-up: ## Start only the database
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d postgres

db-down: ## Stop the database
	docker-compose -f $(DOCKER_COMPOSE_FILE) stop postgres

db-reset: ## Reset the database (WARNING: This will delete all data)
	docker-compose -f $(DOCKER_COMPOSE_FILE) down -v
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d postgres

# Data loading commands
load-initial-data: ## Load initial team data into database
	@echo "📊 Loading initial data..."
	@if [ ! -f scripts/load_initial_data.go ]; then \
		echo "❌ Initial data script not found at scripts/load_initial_data.go"; \
		exit 1; \
	fi
	@echo "Building data loader..."
	@$(GOBUILD) -o scripts/load_initial_data scripts/load_initial_data.go
	@echo "Running data loader..."
	@./scripts/load_initial_data
	@rm -f scripts/load_initial_data
	@echo "✅ Initial data loading completed"

# Migration commands (requires golang-migrate)
migrate-install: ## Install golang-migrate tool
	@if ! command -v migrate > /dev/null; then \
		echo "Installing golang-migrate..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	else \
		echo "golang-migrate is already installed"; \
	fi

migrate-create: ## Create a new migration (usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then \
		echo "Please provide a migration name: make migrate-create NAME=your_migration_name"; \
		exit 1; \
	fi
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(NAME)

migrate-up: ## Run all pending migrations
	migrate -path $(MIGRATION_DIR) -database "postgres://postgres:postgres@localhost:5432/developer_portal?sslmode=disable" up

migrate-down: ## Rollback the last migration
	migrate -path $(MIGRATION_DIR) -database "postgres://postgres:postgres@localhost:5432/developer_portal?sslmode=disable" down 1

migrate-force: ## Force migration version (usage: make migrate-force VERSION=version_number)
	@if [ -z "$(VERSION)" ]; then \
		echo "Please provide a version: make migrate-force VERSION=version_number"; \
		exit 1; \
	fi
	migrate -path $(MIGRATION_DIR) -database "postgres://postgres:postgres@localhost:5432/developer_portal?sslmode=disable" force $(VERSION)

migrate-version: ## Show current migration version
	migrate -path $(MIGRATION_DIR) -database "postgres://postgres:postgres@localhost:5432/developer_portal?sslmode=disable" version

# Code quality commands
lint: ## Run linter (requires golangci-lint)
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

format: ## Format code
	$(GOCMD) fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	else \
		echo "goimports not found. Install it with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

# Setup commands
setup: deps migrate-install ## Setup development environment
	@echo "Setting up development environment..."
	@echo "1. Installing dependencies..."
	@$(MAKE) deps
	@echo "2. Installing migration tool..."
	@$(MAKE) migrate-install
	@echo "3. Starting database..."
	@$(MAKE) db-up
	@echo "4. Waiting for database to be ready..."
	@sleep 10
	@echo "5. Running migrations..."
	@$(MAKE) migrate-up || echo "No migrations to run or database not ready"
	@echo "✅ Setup complete!"
	@echo ""
	@echo "🚀 Ready to start:"
	@echo "   make run-dev        # Development mode (resets DB & loads data)"
	@echo "   make run-dev-quick  # Development mode (keeps existing DB)"
	@echo "   make run            # Production mode (requires JWT_SECRET env var)"

# Environment file
env: ## Create .env file from example
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo ".env file created from .env.example"; \
	else \
		echo ".env file already exists"; \
	fi

# Full development workflow
start: setup env ## Complete setup and start development
	@$(MAKE) dev

# Production commands
build-prod: ## Build for production
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -installsuffix cgo -ldflags '-w -s' -o $(BINARY_NAME) ./cmd/server

# Cleanup commands
clean-all: clean docker-down ## Clean everything
	docker system prune -f
	docker volume prune -f

# ========================================================================================
# TESTING COMMANDS
# ========================================================================================

# Test Database Commands
test-db-up: ## Start test database with Docker
	docker-compose -f docker/docker-compose.test.yml up -d postgres-test
	@echo "Waiting for test database to be ready..."
	@sleep 10

test-db-down: ## Stop test database
	docker-compose -f docker/docker-compose.test.yml down

test-db-reset: ## Reset test database (WARNING: This will delete all test data)
	docker-compose -f docker/docker-compose.test.yml down -v
	docker-compose -f docker/docker-compose.test.yml up -d postgres-test
	@echo "Waiting for test database to be ready..."
	@sleep 10

# Test Setup and Teardown
test-setup: test-db-up ## Setup test environment
	@echo "Test environment setup complete"

test-teardown: test-db-down ## Teardown test environment
	@echo "Test environment torn down"

# Unit Tests (fast, no external dependencies)
test-unit: ## Run unit tests only
	$(GOTEST) -short -v ./internal/service/... ./internal/testutils/...

# Integration Tests (with real database)
test-integration: test-setup ## Run integration tests with real database
	@echo "Running integration tests..."
	TEST_DATABASE_URL="postgres://testuser:testpass@localhost:5433/testdb?sslmode=disable" \
	$(GOTEST) -v ./internal/repository/... ./internal/api/handlers/...

# All Tests
test-all: test-setup ## Run all tests (unit + integration)
	@echo "Running all tests..."
	TEST_DATABASE_URL="postgres://testuser:testpass@localhost:5433/testdb?sslmode=disable" \
	$(GOTEST) -v ./...

# Test with Coverage
test-coverage-full: test-setup ## Run all tests with coverage report
	@echo "Running tests with coverage..."
	TEST_DATABASE_URL="postgres://testuser:testpass@localhost:5433/testdb?sslmode=disable" \
	$(GOTEST) -v -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Test with Race Detection
test-race: test-setup ## Run tests with race detection
	@echo "Running tests with race detection..."
	TEST_DATABASE_URL="postgres://testuser:testpass@localhost:5433/testdb?sslmode=disable" \
	$(GOTEST) -race -v ./...

# Benchmark Tests
test-bench: test-setup ## Run benchmark tests
	@echo "Running benchmark tests..."
	TEST_DATABASE_URL="postgres://testuser:testpass@localhost:5433/testdb?sslmode=disable" \
	$(GOTEST) -bench=. -benchmem ./...

# Verbose Tests
test-verbose: test-setup ## Run tests with verbose output
	@echo "Running tests with verbose output..."
	TEST_DATABASE_URL="postgres://testuser:testpass@localhost:5433/testdb?sslmode=disable" \
	$(GOTEST) -v -count=1 ./...

# Docker Tests (run tests inside Docker container)
test-docker: ## Run tests inside Docker container
	docker-compose -f docker/docker-compose.test.yml up --build --abort-on-container-exit test-runner

test-docker-up: ## Start test environment with Docker
	docker-compose -f docker/docker-compose.test.yml up -d

test-docker-down: ## Stop Docker test environment
	docker-compose -f docker/docker-compose.test.yml down

# Test Cleanup
test-clean: test-teardown ## Clean test artifacts and environment
	rm -f coverage.out coverage.html
	docker-compose -f docker/docker-compose.test.yml down -v
	docker system prune -f

# Mock Generation
mocks: ## Generate mocks for testing (requires gomock)
	@echo "Generating mocks..."
	@if ! command -v mockgen > /dev/null; then \
		echo "Installing gomock..."; \
		go install github.com/golang/mock/mockgen@latest; \
	fi
	@echo "Generating repository mocks..."
	@mkdir -p mocks/repository
	mockgen -source=internal/repository/organization.go -destination=mocks/repository/organization.go
	mockgen -source=internal/repository/member.go -destination=mocks/repository/member.go
	mockgen -source=internal/repository/team.go -destination=mocks/repository/team.go
	mockgen -source=internal/repository/project.go -destination=mocks/repository/project.go
	mockgen -source=internal/repository/component.go -destination=mocks/repository/component.go
	mockgen -source=internal/repository/landscape.go -destination=mocks/repository/landscape.go
	mockgen -source=internal/repository/component_deployment.go -destination=mocks/repository/component_deployment.go
	@echo "Generating service mocks..."
	@mkdir -p mocks/service
	mockgen -source=internal/service/organization.go -destination=mocks/service/organization.go
	mockgen -source=internal/service/member.go -destination=mocks/service/member.go
	mockgen -source=internal/service/team.go -destination=mocks/service/team.go
	mockgen -source=internal/service/project.go -destination=mocks/service/project.go
	mockgen -source=internal/service/component.go -destination=mocks/service/component.go
	mockgen -source=internal/service/landscape.go -destination=mocks/service/landscape.go
	mockgen -source=internal/service/component_deployment.go -destination=mocks/service/component_deployment.go
	@echo "Mocks generated successfully!"

# Quick Test (for development workflow)
test-quick: ## Quick test run (unit tests only)
	$(GOTEST) -short ./...

# Pre-commit Tests (recommended before committing)
test-precommit: format lint test-unit ## Run pre-commit tests (format, lint, unit tests)
	@echo "All pre-commit checks passed!"

# CI/CD Test Pipeline
test-ci: test-setup test-coverage-full test-race ## Full CI/CD test pipeline
	@echo "CI/CD test pipeline completed successfully!"

# Test Help
test-help: ## Show testing help
	@echo "Testing Commands Overview:"
	@echo ""
	@echo "Quick Development:"
	@echo "  test-quick      - Fast unit tests only"
	@echo "  test-precommit  - Pre-commit checks (format, lint, unit tests)"
	@echo ""
	@echo "Comprehensive Testing:"
	@echo "  test-all        - All tests (unit + integration)"
	@echo "  test-coverage-full - Tests with coverage report"
	@echo "  test-race       - Tests with race detection"
	@echo ""
	@echo "Environment Management:"
	@echo "  test-setup      - Setup test database"
	@echo "  test-teardown   - Cleanup test database"
	@echo "  test-clean      - Full cleanup"
	@echo ""
	@echo "Docker Testing:"
	@echo "  test-docker     - Run tests in Docker container"
	@echo "  test-docker-up  - Start Docker test environment"
	@echo ""
	@echo "Development Tools:"
	@echo "  mocks          - Generate test mocks"
	@echo "  test-verbose   - Verbose test output"
	@echo "  test-bench     - Benchmark tests"
