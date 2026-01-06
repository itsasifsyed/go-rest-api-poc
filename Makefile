.PHONY: help migration migrate seed start stop build clean test

# Default target
help:
	@echo "Available commands:"
	@echo "  make migration name=<name>  - Create new migration files"
	@echo "  make migrate                - Run all pending migrations"
	@echo "  make seed                   - Run database seeds"
	@echo "  make start                  - Start the application"
	@echo "  make stop                   - Stop the application"
	@echo "  make build                  - Build the application"
	@echo "  make clean                  - Clean build artifacts"
	@echo "  make test                   - Run tests"
	@echo ""
	@echo "Examples:"
	@echo "  make migration name=add_user_phone"
	@echo "  make migrate"
	@echo "  make start"

# Create a new migration
migration:
	@if [ -z "$(name)" ]; then \
		echo "Error: name is required. Usage: make migration name=add_user_phone"; \
		exit 1; \
	fi
	@NEXT=$$(ls internal/infra/db/migrations/*.up.sql 2>/dev/null | wc -l | xargs); \
	NEXT=$$((NEXT + 1)); \
	PADDED=$$(printf "%06d" $$NEXT); \
	UP_FILE="internal/infra/db/migrations/$${PADDED}_$(name).up.sql"; \
	DOWN_FILE="internal/infra/db/migrations/$${PADDED}_$(name).down.sql"; \
	echo "-- Migration: $(name)" > $$UP_FILE; \
	echo "-- Add your SQL here" >> $$UP_FILE; \
	echo "" >> $$UP_FILE; \
	echo "-- Migration: $(name)" > $$DOWN_FILE; \
	echo "-- Add your rollback SQL here" >> $$DOWN_FILE; \
	echo "" >> $$DOWN_FILE; \
	echo "Created migration files:"; \
	echo "  $$UP_FILE"; \
	echo "  $$DOWN_FILE"

# Run migrations
migrate:
	@echo "Running migrations..."
	@go run cmd/migrate/main.go up

# Run seeds
seed:
	@echo "Running seeds..."
	@echo "Note: Seeds run automatically with 'make start' in non-production environments"
	@echo "To run seeds, start the application: make start"

# Start the application
start:
	@echo "Starting application..."
	@go run cmd/api/main.go

# Stop the application (if running in background)
stop:
	@echo "Stopping application..."
	@pkill -f "cmd/api/main.go" || echo "No running application found"

# Build the application
build:
	@echo "Building application..."
	@go build -o bin/api cmd/api/main.go
	@echo "Binary created at: bin/api"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	@go test ./... -v

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

# Run the application in background
start-bg:
	@echo "Starting application in background..."
	@nohup go run cmd/api/main.go > app.log 2>&1 &
	@echo "Application started. Logs: app.log"
	@echo "PID: $$(pgrep -f 'cmd/api/main.go')"

# View application logs (if running in background)
logs:
	@tail -f app.log

# Database commands
db-reset:
	@echo "⚠️  WARNING: This will delete ALL data!"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		psql $$DB_CONNECTION_STRING -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;" && \
		echo "Database reset complete. Run 'make start' to recreate schema."; \
	else \
		echo "Cancelled."; \
	fi

# Check database status
db-status:
	@echo "Checking database status..."
	@psql $$DB_CONNECTION_STRING -c "SELECT * FROM schema_migrations;" 2>/dev/null || echo "No migrations applied yet"

# Create a new seed file
create-seed:
	@if [ -z "$(name)" ]; then \
		echo "Error: name is required. Usage: make create-seed name=categories"; \
		exit 1; \
	fi
	@NEXT=$$(ls internal/infra/db/seeds/*.sql 2>/dev/null | wc -l | xargs); \
	NEXT=$$((NEXT + 1)); \
	PADDED=$$(printf "%03d" $$NEXT); \
	SEED_FILE="internal/infra/db/seeds/$${PADDED}_$(name).sql"; \
	echo "-- Seed: $(name)" > $$SEED_FILE; \
	echo "-- Add your seed data here with ON CONFLICT DO NOTHING" >> $$SEED_FILE; \
	echo "INSERT INTO table_name (id, name) VALUES" >> $$SEED_FILE; \
	echo "    ('id-1', 'value-1')" >> $$SEED_FILE; \
	echo "ON CONFLICT (id) DO NOTHING;" >> $$SEED_FILE; \
	echo "" >> $$SEED_FILE; \
	echo "Created seed file: $$SEED_FILE"

# Development mode with auto-reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	@air

