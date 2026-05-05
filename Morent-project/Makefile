# Makefile для Morent

DOCKER_COMPOSE=docker compose
COMPOSE_PROD_ARGS=-f docker-compose.yml -f docker-compose.prod.override.yml

.PHONY: build up down restart logs backend seed frontend up-prod down-prod restart-prod build-prod clean lint test

# Сборка Docker-образов
build:
	$(DOCKER_COMPOSE) build

# Запуск docker-compose
up:
	$(DOCKER_COMPOSE) up -d --build

# Запуск в режиме "прод": docker-compose.yml + docker-compose.prod.override.yml
up-prod:
	$(DOCKER_COMPOSE) $(COMPOSE_PROD_ARGS) up -d --build

# Остановка и удаление контейнеров
# (останавливает и удаляет все сервисы)
down:
	$(DOCKER_COMPOSE) down

down-prod:
	$(DOCKER_COMPOSE) $(COMPOSE_PROD_ARGS) down

# Перезапуск (down + up)
restart:
	$(MAKE) down
	$(MAKE) up

# Жесткий перезапуск
restart-force:
	$(DOCKER_COMPOSE) down
	$(DOCKER_COMPOSE) up -d --build

# Перезапуск в прод-режиме
restart-prod:
	$(MAKE) down-prod
	$(MAKE) up-prod

# Логи всех сервисов
logs:
	$(DOCKER_COMPOSE) logs -f

# Чистка контейнеров и volumes
clean:
	$(DOCKER_COMPOSE) down -v --remove-orphans

# Базовая проверка: тесты и линтеры через go toolchain
lint:
	cd backend && go test ./...

test:
	cd backend && go test ./...

# Запуск бэкенда (БЕЗ пересида данных)
backend:
	cd backend && go run cmd/server/main.go

# Явный запуск сидов (ОДНОРАЗОВО или по необходимости)
seed:
	cd backend && go run cmd/server/main.go seed

frontend:
	cd frontend && npm run dev

# Производственные сборки (для CI/деплоя)
build-backend:
	cd backend && go build ./cmd/server

build-frontend:
	cd frontend && npm run build

build-prod:
	$(MAKE) build-backend
	$(MAKE) build-frontend

