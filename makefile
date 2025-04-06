# Makefile for managing migrations, sqlc, and Go build/deploy

GO_BINARY=go
DB_URL=$(shell echo $$DB_URL)  

MIGRATIONS_PATH=migrations
SQLC_PATH=db/sqlc


GEN_CODE_DIR=./db/sqlc

# Go project settings
GO_FLAGS=-v

# Migration tool settings 
MIGRATE=migrate
MIGRATE_CMD=$(MIGRATE) -path $(MIGRATIONS_PATH) -database $(DB_URL)



## Run all migrations (up)
migrate-up:
	$(MIGRATE_CMD) up

## Rollback the last applied migration (down)
migrate-down:
	$(MIGRATE_CMD) down 1

## Create a new migration file
migrate-create:
	@echo "Enter migration name: "
	@read name && \
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_PATH) $$name

## Show the current status of migrations
migrate-status:
	$(MIGRATE_CMD) version

# Generate Go code with sqlc based on SQL queries and schema
sqlc-generate:
	sqlc generate

# Build the Go application
build:
	$(GO_BINARY) build $(GO_FLAGS) -o app


# Run the application
run: build
	./app

# Clean up the build artifacts
clean:
	$(GO_BINARY) clean
	rm -f ./app

# Install all Go dependencies
deps:
	$(GO_BINARY) mod tidy


# Rebuild and run the app, ensuring everything is up-to-date
rebuild: clean build


# Full setup for initial setup or CI
setup: deps install-sqlc install-migrate

# Command to run the full pipeline of migrations and then generate sqlc code
migrate-and-generate:
	$(MAKE) migrate-up
	$(MAKE) sqlc-generate

.PHONY: migrate-up migrate-down migrate-create migrate-status sqlc-generate build test run clean deps install-sqlc install-migrate rebuild setup migrate-and-generate
