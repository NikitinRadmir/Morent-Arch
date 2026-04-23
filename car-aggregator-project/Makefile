.PHONY: build run test test-integration clean docker-build docker-run help

# Default target
help:
	@echo "Available commands:"
	@echo "  build            - Build the application"
	@echo "  run              - Run the application locally"
	@echo "  test             - Run unit tests"
	@echo "  test-integration - Run integration tests (requires API keys)"
	@echo "  test-all         - Run all tests"
	@echo "  clean            - Clean build artifacts"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-run       - Run application in Docker"
	@echo "  lint             - Run linter"

# Build the application
build:
	go build -o bin/car-aggregator .

# Run the application locally
run:
	go run main.go

# Run unit tests
test:
	go test -v ./internal/...

# Run integration tests (requires API keys in .env)
test-integration:
	go test -v ./integration/...

# Run all tests
test-all: test test-integration

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Build Docker image
docker-build:
	docker build -t car-aggregator .

# Run application in Docker
docker-run:
	docker-compose up --build

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Download dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...