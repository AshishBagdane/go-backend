.PHONY: help build run test clean install deps swagger swagger-install swagger-validate check format lint db-migrate db-status db-rollback db-driver bootstrap docker-build docker-run

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download Go dependencies
	go mod download
	go mod tidy

build: swagger ## Build the application (regenerates swagger docs first)
	go build -o bin/go-backend cmd/server/main.go

run: ## Run the application
	go run cmd/server/main.go

test: ## Run API tests (requires server to be running)
	./test_api.sh

install: deps ## Install dependencies and build
	@echo "Installing dependencies..."
	go mod download
	@echo "Building application..."
	$(MAKE) build
	@echo "Done! Run 'make run' to start the server"

clean: ## Clean build artifacts
	rm -rf bin/
	go clean

swagger-install: ## Install swag CLI tool
	@which swag > /dev/null 2>&1 || (echo "Installing swag..." && go install github.com/swaggo/swag/cmd/swag@latest)

swagger: swagger-install ## Regenerate Swagger docs from annotations
	swag init -g cmd/server/main.go
	@echo "Swagger docs regenerated at docs/swagger.json"

swagger-validate: swagger ## Regenerate and validate Swagger docs
	@which swagger > /dev/null 2>&1 && swagger validate docs/swagger.json || echo "(install 'swagger' CLI to validate the spec)"

docker-build: ## Build Docker image
	docker build -t go-backend:latest .

docker-run: ## Run Docker container
	docker run -p 8080:8080 --env-file .env go-backend:latest

format: ## Format Go code
	go fmt ./...

lint: ## Run Go linter
	golangci-lint run || go vet ./...

db-status: ## Check Liquibase migration status
	@echo "Checking database migration status..."
	@liquibase status

db-driver: ## Download Postgres JDBC driver for Liquibase (optional)
	@mkdir -p tools
	@echo "Downloading Postgres JDBC driver..."
	@curl -L -o tools/postgresql.jar https://repo1.maven.org/maven2/org/postgresql/postgresql/42.7.4/postgresql-42.7.4.jar

db-migrate: ## Run pending Liquibase migrations
	@echo "Running database migrations..."
	@liquibase update

db-rollback: ## Rollback the last Liquibase migration (requires 1 argument, e.g. make db-rollback COUNT=1)
	@if [ -z "$(COUNT)" ]; then \
		echo "Error: COUNT is required. Usage: make db-rollback COUNT=1"; \
		exit 1; \
	fi
	@echo "Rolling back $(COUNT) migration(s)..."
	@liquibase rollbackCount $(COUNT)

check: ## Run all checks (format, lint, swagger)
	$(MAKE) format
	$(MAKE) lint
	$(MAKE) swagger

bootstrap: deps db-driver db-migrate swagger run ## One-shot local setup and run
