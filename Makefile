# Общий Makefile — сборка и запуск микросервисов Morent + Kafka
# Запуск из корня репозитория: make up

.DEFAULT_GOAL := help

DOCKER         ?= docker compose
ROOT_DIR       := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))

KAFKA_FILE     := $(ROOT_DIR)/infra/kafka/docker-compose.yml
KAFKA_PROJECT  := kafka
OBS_FILE       := $(ROOT_DIR)/infra/observability/docker-compose.yml
OBS_PROJECT    := morent-obs

MORENT_DIR     := $(ROOT_DIR)/Morent-project
USER_DIR       := $(ROOT_DIR)/user-system-develop
AGGREGATOR_DIR := $(ROOT_DIR)/car-aggregator-project
PAYMENT_DIR    := $(ROOT_DIR)/payment-service
GENERATOR_DIR  := $(ROOT_DIR)/generator-service
EMAIL_DIR      := $(ROOT_DIR)/EmailService
EMAIL_PROJECT  := $(EMAIL_DIR)/EmailService.Api
EMAIL_FILE     := $(EMAIL_PROJECT)/docker-compose.yml
EMAILTEST_DIR  := $(ROOT_DIR)/emailtest

# --- Kafka (инфра, общая сеть morent-kafka → kafka_morent-kafka) ---

.PHONY: kafka-up kafka-down kafka-logs kafka-ps kafka-build \
	obs-up obs-down obs-logs obs-ps

obs-up:
	$(DOCKER) -f $(OBS_FILE) -p $(OBS_PROJECT) up -d

obs-down:
	$(DOCKER) -f $(OBS_FILE) -p $(OBS_PROJECT) down

obs-logs:
	$(DOCKER) -f $(OBS_FILE) -p $(OBS_PROJECT) logs -f

obs-ps:
	$(DOCKER) -f $(OBS_FILE) -p $(OBS_PROJECT) ps

kafka-up:
	$(DOCKER) -f $(KAFKA_FILE) -p $(KAFKA_PROJECT) up -d

kafka-down:
	$(DOCKER) -f $(KAFKA_FILE) -p $(KAFKA_PROJECT) down

kafka-logs:
	$(DOCKER) -f $(KAFKA_FILE) -p $(KAFKA_PROJECT) logs -f

kafka-ps:
	$(DOCKER) -f $(KAFKA_FILE) -p $(KAFKA_PROJECT) ps

kafka-build:
	$(DOCKER) -f $(KAFKA_FILE) -p $(KAFKA_PROJECT) pull

# --- Сборка образов (без запуска) ---

.PHONY: build build-morent build-user build-aggregator build-payment build-generator build-email

build: build-morent build-user build-aggregator build-payment build-generator build-email

build-morent:
	$(DOCKER) --project-directory $(MORENT_DIR) -f $(MORENT_DIR)/docker-compose.yml build

build-user:
	$(DOCKER) --project-directory $(USER_DIR) -f $(USER_DIR)/docker-compose.yml build

build-aggregator:
	$(DOCKER) --project-directory $(AGGREGATOR_DIR) -f $(AGGREGATOR_DIR)/docker-compose.yml build

build-payment:
	$(DOCKER) --project-directory $(PAYMENT_DIR) -f $(PAYMENT_DIR)/docker-compose.yml build

build-generator:
	$(DOCKER) --project-directory $(GENERATOR_DIR) -f $(GENERATOR_DIR)/docker-compose.yml build

build-email:
	$(DOCKER) --project-directory $(EMAIL_PROJECT) -f $(EMAIL_FILE) build

# --- Запуск / остановка отдельных сервисов ---

.PHONY: morent-up morent-down morent-build morent-logs \
	user-up user-down user-build user-logs \
	aggregator-up aggregator-down aggregator-build aggregator-logs \
	payment-up payment-down payment-build payment-logs \
	generator-up generator-down generator-build generator-logs \
	email-up email-down email-build email-logs \
	emailtest-up emailtest-down emailtest-logs

morent-up:
	$(DOCKER) --project-directory $(MORENT_DIR) -f $(MORENT_DIR)/docker-compose.yml up -d --build

morent-down:
	$(DOCKER) --project-directory $(MORENT_DIR) -f $(MORENT_DIR)/docker-compose.yml down

morent-build: build-morent

morent-logs:
	$(DOCKER) --project-directory $(MORENT_DIR) -f $(MORENT_DIR)/docker-compose.yml logs -f

user-up:
	$(DOCKER) --project-directory $(USER_DIR) -f $(USER_DIR)/docker-compose.yml up -d --build

user-down:
	$(DOCKER) --project-directory $(USER_DIR) -f $(USER_DIR)/docker-compose.yml down

user-build: build-user

user-logs:
	$(DOCKER) --project-directory $(USER_DIR) -f $(USER_DIR)/docker-compose.yml logs -f

aggregator-up:
	$(DOCKER) --project-directory $(AGGREGATOR_DIR) -f $(AGGREGATOR_DIR)/docker-compose.yml up -d --build

aggregator-down:
	$(DOCKER) --project-directory $(AGGREGATOR_DIR) -f $(AGGREGATOR_DIR)/docker-compose.yml down

aggregator-build: build-aggregator

aggregator-logs:
	$(DOCKER) --project-directory $(AGGREGATOR_DIR) -f $(AGGREGATOR_DIR)/docker-compose.yml logs -f

payment-up:
	$(DOCKER) --project-directory $(PAYMENT_DIR) -f $(PAYMENT_DIR)/docker-compose.yml up -d --build

payment-down:
	$(DOCKER) --project-directory $(PAYMENT_DIR) -f $(PAYMENT_DIR)/docker-compose.yml down

payment-build: build-payment

payment-logs:
	$(DOCKER) --project-directory $(PAYMENT_DIR) -f $(PAYMENT_DIR)/docker-compose.yml logs -f

generator-up:
	$(DOCKER) --project-directory $(GENERATOR_DIR) -f $(GENERATOR_DIR)/docker-compose.yml up -d --build

generator-down:
	$(DOCKER) --project-directory $(GENERATOR_DIR) -f $(GENERATOR_DIR)/docker-compose.yml down

generator-build: build-generator

generator-logs:
	$(DOCKER) --project-directory $(GENERATOR_DIR) -f $(GENERATOR_DIR)/docker-compose.yml logs -f

email-up:
	$(DOCKER) --project-directory $(EMAIL_PROJECT) -f $(EMAIL_FILE) up -d --build

email-down:
	$(DOCKER) --project-directory $(EMAIL_PROJECT) -f $(EMAIL_FILE) down

email-build: build-email

email-logs:
	$(DOCKER) --project-directory $(EMAIL_PROJECT) -f $(EMAIL_FILE) logs -f

emailtest-up:
	$(DOCKER) --project-directory $(EMAILTEST_DIR) -f $(EMAILTEST_DIR)/docker-compose.yml up -d --build

emailtest-down:
	$(DOCKER) --project-directory $(EMAILTEST_DIR) -f $(EMAILTEST_DIR)/docker-compose.yml down

emailtest-logs:
	$(DOCKER) --project-directory $(EMAILTEST_DIR) -f $(EMAILTEST_DIR)/docker-compose.yml logs -f

# --- Весь стек ---

.PHONY: up down restart ps logs

# Порядок: Kafka → user-system (consumer) → Morent (producer) → email → остальные
up: kafka-up user-up morent-up email-up aggregator-up payment-up generator-up
	@echo ""
	@echo "Стек поднят."
	@echo "  Kafka UI:     http://localhost:8090"
	@echo "  Kibana:       http://localhost:5601  (после make obs-up)"
	@echo "  Morent UI:    http://localhost:$${FRONTEND_PORT:-5173}"
	@echo "  Morent API:   http://localhost:$${BACKEND_PORT:-1488}"
	@echo "  Email API:    http://localhost:$${EMAIL_API_PORT:-5112}"

down: generator-down payment-down aggregator-down email-down morent-down user-down kafka-down

restart: down up

ps:
	@echo "=== Kafka ==="
	@$(DOCKER) -f $(KAFKA_FILE) -p $(KAFKA_PROJECT) ps
	@echo ""
	@echo "=== Morent ==="
	@$(DOCKER) --project-directory $(MORENT_DIR) -f $(MORENT_DIR)/docker-compose.yml ps
	@echo ""
	@echo "=== user-system ==="
	@$(DOCKER) --project-directory $(USER_DIR) -f $(USER_DIR)/docker-compose.yml ps
	@echo ""
	@echo "=== car-aggregator ==="
	@$(DOCKER) --project-directory $(AGGREGATOR_DIR) -f $(AGGREGATOR_DIR)/docker-compose.yml ps
	@echo ""
	@echo "=== payment-service ==="
	@$(DOCKER) --project-directory $(PAYMENT_DIR) -f $(PAYMENT_DIR)/docker-compose.yml ps
	@echo ""
	@echo "=== generator-service ==="
	@$(DOCKER) --project-directory $(GENERATOR_DIR) -f $(GENERATOR_DIR)/docker-compose.yml ps
	@echo ""
	@echo "=== EmailService ==="
	@$(DOCKER) --project-directory $(EMAIL_DIR) -f $(EMAIL_DIR)/docker-compose.yml ps

logs:
	@echo "Логи Kafka (Ctrl+C для выхода)..."
	$(DOCKER) -f $(KAFKA_FILE) -p $(KAFKA_PROJECT) logs -f

# --- Справка ---

.PHONY: help

help:
	@echo "Morent Architecture — общие команды Docker"
	@echo ""
	@echo "Стек целиком:"
	@echo "  make up        — Kafka + все сервисы (сборка и запуск)"
	@echo "  make down      — остановить всё"
	@echo "  make restart   — down + up"
	@echo "  make build     — собрать образы без запуска"
	@echo "  make ps        — статус контейнеров"
	@echo ""
	@echo "Kafka (infra/kafka):"
	@echo "  make kafka-up / kafka-down / kafka-logs / kafka-ps"
	@echo ""
	@echo "Observability (Elasticsearch, Kibana, Filebeat, Heartbeat):"
	@echo "  make obs-up / obs-down / obs-logs / obs-ps"
	@echo ""
	@echo "Отдельные сервисы (примеры):"
	@echo "  make morent-up      make user-up"
	@echo "  make aggregator-up  make payment-up"
	@echo "  make generator-up   make email-up   make emailtest-up"
	@echo ""
	@echo "Перед первым запуском создайте .env из .env.example в:"
	@echo "  Morent-project, user-system-develop, car-aggregator-project, EmailService"
