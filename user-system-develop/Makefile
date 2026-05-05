ifneq (,$(wildcard .env))
    include .env
    export
endif

APP_NAME = user-system-api
BINARY_NAME = main
GO_MAIN_PATH = ./app
GO_FILES = $(shell find ./app -name "*.go" -type f)

# Docker
DOCKER_COMPOSE = docker-compose.yml
DOCKER_APP_CONTAINER = user-system-api
DOCKER_DB_CONTAINER = user-system-db
DOCKER_RABBITMQ_CONTAINER = user-system-rabbitmq

# PostgreSQL
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= user_system_db
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_SSL ?= disable

# RabbitMQ
RABBITMQ_USER ?= guest
RABBITMQ_PASSWORD ?= guest
RABBITMQ_HOST ?= localhost
RABBITMQ_PORT ?= 5672

# Приложение
APP_PORT ?= 8080

build:
	cd app && go build -o main.exe .

# Запуск
run:
	cd app && go run main.go

# Docker
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Очистка
clean:
	rm -f app/main.exe

# Помощь (упрощенная)
help:
	@echo "Доступные команды:"
	@echo "  make build      - собрать приложение"
	@echo "  make run        - запустить приложение"
	@echo "  make docker-up  - запустить Docker"
	@echo "  make docker-down- остановить Docker"
	@echo "  make clean      - очистить"

.DEFAULT_GOAL := help