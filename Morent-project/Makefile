# Makefile для Morent

DOCKER_COMPOSE=docker-compose

.PHONY: build up down restart logs backend seed frontend

# Сборка Docker-образов
build:
	$(DOCKER_COMPOSE) build

# Запуск docker-compose
up:
	$(DOCKER_COMPOSE) up -d

# Остановка и удаление контейнеров
# (останавливает и удаляет все сервисы)
down:
	$(DOCKER_COMPOSE) down

# Перезапуск (down + up)
restart:
	$(MAKE) down
	$(MAKE) up

# Логи всех сервисов
logs:
	$(DOCKER_COMPOSE) logs -f

# Запуск бэкенда (БЕЗ пересида данных)
backend:
	cd backend && go run cmd/server/main.go

# Явный запуск сидов (ОДНОРАЗОВО или по необходимости)
seed:
	cd backend && go run cmd/server/main.go seed

frontend:
	cd frontend && npm run dev

