# Variables
PROTO_DIR := internal/server/proto
PROTO_FILE := $(PROTO_DIR)/server.proto
GO_PROTO_OUT := internal/server
SERVER_CMD := cmd/server
CLIENT_CMD := cmd/client
BIN_DIR := bin
ENV_FILE := .env

# Default target
.PHONY: all
all: build

# Install dependencies
.PHONY: deps
deps:
	go mod download
	@echo "Dependencies installed"

# Generate gRPC code from proto file
.PHONY: proto
proto:
	@echo "Generating gRPC code..."
	cd $(PROTO_DIR) && \
	protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	server.proto
	@echo "gRPC code generated successfully"

# Build server binary
.PHONY: build-server
build-server:
	@echo "Building server..."
	mkdir -p $(BIN_DIR)
	cd $(SERVER_CMD) && go build -o ../../$(BIN_DIR)/server
	@echo "Server built: $(BIN_DIR)/server"

# Build client binary with version info
.PHONY: build-client
build-client:
	@echo "Building client with version info..."
	mkdir -p $(BIN_DIR)
	cd cmd/client && go build \
		-ldflags="\
		-X 'github.com/dangerousmonk/gophkeeper/internal/version.BuildDate=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")' \
		-X 'github.com/dangerousmonk/gophkeeper/internal/version.GitCommit=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")' \
		-X 'github.com/dangerousmonk/gophkeeper/internal/version.GoVersion=$(shell go version | cut -d" " -f3)'" \
		-o ../../$(BIN_DIR)/client
	@echo "Client built: $(BIN_DIR)/client"

# Build both server and client
.PHONY: build
build: build-server build-client

# Start test database container
.PHONY: db-up
db-up:
	@if [ -f $(ENV_FILE) ]; then \
		echo "Starting database with environment file..."; \
		docker-compose --env-file $(ENV_FILE) up -d; \
	else \
		echo "Starting database without environment file..."; \
		docker-compose up -d; \
	fi
	@echo "Database container started"

# Stop test database container
.PHONY: db-down
db-down:
	docker-compose down
	@echo "Database container stopped"

# Run tests without cache
.PHONY: test
test:
	go test ./... --count=1
	@echo "Tests completed"

# Run tests with coverage and generate HTML report
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./... --count=1
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"


# Generate coverage percentage with formatted output
.PHONY: coverage-percent
coverage-percent:
	@go test -coverprofile=coverage.out ./... > /dev/null 2>&1
	@coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}'); \
	echo "Coverage percent is $$coverage"


# Help message
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make proto       - Generate gRPC code from proto file"
	@echo "  make build-server - Build server binary"
	@echo "  make build-client - Build client binary"
	@echo "  make build       - Build both server and client"
	@echo "  make db-up       - Start test database container"
	@echo "  make db-down     - Stop test database container"
	@echo "  make deps        - Install dependencies"
	@echo "  make test        - Run tests without cache"
	@echo "  make test-coverage - Run tests with coverage"
	@echo "  make coverage-percent - See output coverage percent"
	@echo "  make help        - Show this help message"