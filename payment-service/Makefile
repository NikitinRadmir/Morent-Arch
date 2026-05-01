.PHONY: run build test lint docker-build docker-up docker-down

## Run locally (requires Go 1.22+)
run:
	go run ./cmd/payments

## Build binary to bin/server
build:
	go build -o bin/server ./cmd/payments

## Run unit tests
test:
	go test ./...

## Static analysis
lint:
	go vet ./...

## Build Docker image
docker-build:
	docker compose build

## Start all services in detached mode
docker-up:
	docker compose up -d

## Stop all services
docker-down:
	docker compose down

## Copy env template and start (first run)
setup:
	cp -n .env.dist .env || true
	docker compose up -d
