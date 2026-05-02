.PHONY: run build test lint docker-build docker-up docker-down

run:
	go run ./cmd/server

build:
	go build -o bin/payment-service ./cmd/server

test:
	go test ./...

lint:
	go vet ./...

docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down
