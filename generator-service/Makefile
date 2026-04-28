.PHONY: run build test docker-build docker-run clean

run:
	go run cmd/server/main.go

build:
	go build -o bin/generator-service cmd/server/main.go

test:
	go test -v ./...

docker-build:
	docker build -t generator-service:latest .

docker-run:
	docker run -p 8080:8080 generator-service:latest

clean:
	rm -rf bin/
	go clean
