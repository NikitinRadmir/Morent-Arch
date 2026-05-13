# Morent Project — Car Rental Platform

## Состав проекта

- **backend/** — Backend на Go (REST API, GORM/Postgres, сервисный слой, автогенерация роутов)
- **frontend/** — Frontend на React + Vite
- **docker-compose.yml** — Быстрый запуск БД и (опционально) backend
- **Makefile** — Удобные команды для сборки, запуска, seed, миграций

---

## Быстрый старт

### 1. Запуск через Docker Compose (_рекомендуется_)

- Убедитесь, что установлен Docker Compose
- В корне проекта выполните:

```bash
make build      # Собрать образы
make up         # Запустить контейнеры (Postgres, MinIO если включён)
```

- Установите Minio для хранения ваших логов и других файлов (Minio - объектное хранилище, реализующее Amazon S3):

```https://dl.min.io/server/minio/release/windows-amd64/minio.exe```

- Необходимо перенести exe файл в папку "C:/minio/minio.exe"

-Откройте PowerShell от имени администратора впишите команду:

```C:\minio\minio.exe server C:\minio\data --console-address ":9001"```

- Теперь у вас запущен сервер Minio и логи с данными будут сохраняться туда


- После этого можно запускать backend локально (см. ниже). Убедитесь, что Postgres реально поднялся (см. `make logs`).

### 2. Запуск backend локально

```bash
make backend       # (или cd backend && go run cmd/server/main.go)
```

### 3. Seed тестовых данных (по необходимости)

```bash
make seed          # Заполнить БД тестовыми машинами и отзывами
```

### 4. Запуск фронта

```bash
make frontend      # Откроется на http://localhost:5173
```

---

## gRPC (бронирования / rentals)

Backend поднимает **два сервера в одном процессе**:

- **HTTP (REST + GraphQL + Swagger)**: `http://localhost:1488`
- **gRPC (BookingService)**: `localhost:50051` (порт задаётся переменной окружения `GRPC_PORT`)

### Где лежит контракт

- `.proto`: `backend/proto/booking/v1/booking.proto`

### Методы сервиса

- `CreateBooking`
- `GetBookingById`
- `ListCarBookings` (занятые периоды по авто)
- `CancelBooking`

## Структура backend

- **internal/handlers/** — HTTP handlers (REST контроллеры)
- **internal/repository/** — Репозитории (CRUD/фильтрация через GORM)
- **internal/service/** — Слой бизнес-логики (CarService, CommentService)
- **internal/di/** — Контейнер зависимостей (DI)
- **internal/server/** — Универсальный router с автогенерацией путей
- **internal/migrations/** — Авто-миграции моделей через GORM
- **internal/seed/** — Seed функций для заполнения БД
- **cmd/server/main.go** — Точка входа, DI, роутинг, graceful shutdown, seed

---

## Примеры запросов к backend API

- Получить все машины:         `GET    /Cars/GetAll`
- Получить по id:              `GET    /Cars/GetById/{id}`
- Фильтрация:                  `GET    /Cars/GetFiltered?name=...&type=...`
- Добавить машину:             `POST   /Cars/Create`
- Обновить:                    `PUT    /Cars/Update/{id}`
- Удалить:                     `DELETE /Cars/Delete/{id}`

(Аналогично для /Comments/...)

---

## Важно
- Вся конфигурация backend — **только через переменные окружения** (см. `.env.example` в корне `Morent-project`). Файла `config.json` нет.
- Пароль Postgres в Docker по умолчанию: `postgresmaster`
- Для полного пересоздания базы: `make down` + `docker volume rm morent-project_postgres_data` + потом снова `make up` + `make seed`

---

## Дополнительные команды

```bash
make build    # Пересобрать Docker-образы
make up       # Запустить сервисы (БД)
make down     # Остановить
make logs     # Логи контейнеров
make backend  # Локальный запуск Go backend
make frontend # Локальный запуск React фронта
make seed     # (Пере)загрузить тестовые данные в БД
```

---

## Стек и фичи
- Go + net/http + GORM ORM
- Автоматические миграции
- DI контейнер
- Автоматизация роутинга через группы
- Frontend на React (Vite, компонентная архитектура)
- Стиль современный dark/light
- Готов к деплою в Docker

---

## Вопросы/Запуск/Проблемы?
- Проверьте `.env` (скопируйте из `.env.example`) — пароли, порты, `DB_HOST`/`DB_PORT`, MinIO, Redis.
- Остановите локальный Postgres на порту 5432/5433, если конфликтует с Docker
- Для любых вопросов: см. README и комментарии в коде

---

## Обновлённое описание и запуск проекта

### Что делает проект

Morent — это платформа аренды авто с полным циклом:

- публичный каталог машин, страница детали авто;
- бронирование с выбором дат и проверкой пересечений;
- отзывы, которые могут оставлять только пользователи с фактической арендой;
- избранное, личный кабинет пользователя;
- админ‑панель (управление пользователями, арендами, избранным, комментариями и логами);
- бизнес‑логирование действий админки в MinIO и просмотр истории в админке;
- минимальная OpenAPI‑схема (`/swagger.json`) для интеграции со Swagger UI / Postman.

### Запуск по шагам (локальная разработка)

0. **Переменные окружения**: скопируйте `Morent-project/.env.example` в `.env` в том же каталоге. Для `make backend` вне Docker задайте `DB_HOST`, `DB_PORT`, `MINIO_ENDPOINT` и при необходимости `REDIS_ADDR` (см. комментарии в `.env.example`).

1. **Поднять инфраструктуру (Postgres, MinIO по желанию)**:

```bash
make up
```

2. **Прогнать миграции (делаются при старте backend) и, при необходимости, сиды**:

```bash
make backend   # миграции выполняются при старте
make seed      # сиды (по необходимости)
```

3. **Запустить backend** (по умолчанию `http://localhost:1488`):

```bash
make backend
```

4. **Запустить frontend** (по умолчанию `http://localhost:5173`):

```bash
cd frontend
npm install
npm run dev
```

Адрес API задаётся через `VITE_API_URL` (см. `frontend/src/context/AuthContext.jsx`).  
Если переменная не задана, используется `http://localhost:1488`.

### Важные аккаунты и роли

При миграциях автоматически создаётся админ‑пользователь:

- **email**: `admin@morent.com`
- **password**: `admin123`

Роль (`Role`) этого пользователя — `admin`, что даёт доступ к эндпоинтам `/Admin/...` и админ‑панели на фронте.

### Где смотреть API

- Базовый список ручек описан выше в разделе «Примеры запросов к backend API»;
- JSON‑описание OpenAPI доступно по `GET /swagger.json` (можно скормить Swagger UI или Postman);
- бизнес‑логи за текущий день доступны через `GET /Admin/Logs` (только для роли `admin`). 