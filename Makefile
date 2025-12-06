.PHONY: help dev-up dev-down dev-restart dev-logs db-migrate db-reset minio-console test clean

# Load environment variables
include .env
export

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev-up: ## Start all development services (PostgreSQL + MinIO)
	@echo "🚀 Starting development services..."
	docker-compose up -d
	@echo "✅ Services started!"
	@echo ""
	@echo "📊 Service URLs:"
	@echo "  - PostgreSQL:    localhost:5432"
	@echo "  - MinIO API:     http://localhost:9000"
	@echo "  - MinIO Console: http://localhost:9001"
	@echo "  - Adminer:       http://localhost:8081 (run 'make tools-up' to enable)"
	@echo ""
	@echo "🔑 MinIO Credentials:"
	@echo "  - Username: minioadmin"
	@echo "  - Password: minioadmin123"
	@echo ""
	@echo "💡 Run 'make dev-logs' to see logs"

dev-down: ## Stop all development services
	@echo "🛑 Stopping development services..."
	docker-compose down
	@echo "✅ Services stopped!"

dev-restart: dev-down dev-up ## Restart all development services

dev-logs: ## Show logs from all services
	docker-compose logs -f

dev-status: ## Show status of all services
	docker-compose ps

tools-up: ## Start development tools (Adminer)
	docker-compose --profile tools up -d
	@echo "✅ Development tools started!"
	@echo "  - Adminer (DB UI): http://localhost:8081"

tools-down: ## Stop development tools
	docker-compose --profile tools down

db-connect: ## Connect to PostgreSQL with psql
	docker-compose exec postgres psql -U docuser -d docapi

db-migrate: ## Run database migrations
	@echo "🔄 Running database migrations..."
	go run cmd/api/main.go migrate
	@echo "✅ Migrations completed!"

db-reset: ## Reset database (drops and recreates)
	@echo "⚠️  This will delete all data!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose exec postgres psql -U docuser -d postgres -c "DROP DATABASE IF EXISTS docapi;"; \
		docker-compose exec postgres psql -U docuser -d postgres -c "CREATE DATABASE docapi;"; \
		echo "✅ Database reset!"; \
	fi

minio-console: ## Open MinIO console in browser
	@echo "🌐 Opening MinIO console..."
	@open http://localhost:9001 || xdg-open http://localhost:9001 || echo "Please open http://localhost:9001 in your browser"

minio-create-bucket: ## Create MinIO bucket manually
	docker-compose exec minio-client mc mb myminio/documents --ignore-existing
	@echo "✅ Bucket 'documents' created!"

run: ## Run the API server
	@echo "🚀 Starting API server..."
	go run cmd/api/main.go

test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "✅ Tests completed!"
	@echo "📊 Coverage report generated: coverage.out"

test-coverage: test ## Run tests and show coverage in browser
	go tool cover -html=coverage.out

build: ## Build the API binary
	@echo "🔨 Building API..."
	go build -o bin/api cmd/api/main.go
	@echo "✅ Binary created: bin/api"

clean: ## Clean up generated files and volumes
	@echo "🧹 Cleaning up..."
	docker-compose down -v
	rm -rf bin/
	rm -f coverage.out
	@echo "✅ Cleanup completed!"

install-deps: ## Install Go dependencies
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies installed!"

lint: ## Run linter
	@echo "🔍 Running linter..."
	golangci-lint run ./...
	@echo "✅ Linting completed!"

docker-build: ## Build Docker image for the API
	@echo "🐳 Building Docker image..."
	docker build -t document-summarizer-api:latest .
	@echo "✅ Docker image built!"

docker-run: ## Run API in Docker with docker-compose
	@echo "🐳 Running API in Docker..."
	docker-compose -f docker-compose.full.yml up -d
	@echo "✅ API running in Docker!"

.DEFAULT_GOAL := help