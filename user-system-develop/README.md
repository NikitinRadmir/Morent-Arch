# User System API

Система управления пользователями с аутентификацией, ролями и правами доступа.

## Основные функции

- Регистрация и авторизация (JWT)
- Управление пользователями (CRUD)
- Управление ролями и правами доступа (RBAC)
- Управление компаниями
- Синхронизация пользователей из Morent через Kafka
- RPC вызовы

## API Документация

OpenAPI файлы в папке `/docs`:
- `openapi-main.yaml` - основная документация
- `openapi-auth.yaml` - аутентификация
- `openapi-users.yaml` - пользователи
- `openapi-roles.yaml` - роли
- `openapi-companies.yaml` - компании

Просмотр документации:
```bash
# Swagger UI
docker run -p 8081:8080 -v ${PWD}/docs:/usr/share/nginx/html/docs swaggerapi/swagger-ui
```

## Сборка и запуск

### 1. Клонирование репозитория

git clone https://git.kpfu.ru/2025-golang/projects/user-system.git
cd user-system

### 2. Настройка окружения

cp .env.example .env
При необходимости отредактируйте .env

### 3. Сборка и запуск

## Способ 1: Docker (рекомендуется)
### 1. Собрать образы и запустить контейнеры
docker-compose up -d --build
### 2. Проверить статус
docker-compose ps
### 3. Посмотреть логи
docker-compose logs -f
### 4. Остановить
docker-compose down

## Способ 2: Makefile
### 1. Собрать приложение
make build
### 2. Запустить Docker контейнеры
make docker-up
### 3. Запустить локально (требуется PostgreSQL; для Kafka — `infra/kafka`)
make run
### 4. Остановить
make docker-down