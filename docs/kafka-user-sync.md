# Синхронизация пользователей Morent → user-system (Kafka)

Интеграция **только через Kafka**. Morent не вызывает user-system по HTTP при регистрации, логине или смене профиля.

## Схема

1. Пользователь работает с **Morent** (`POST /auth/register`, логин, профиль, пароль).
2. Morent сохраняет пользователя в свою БД и публикует событие в Kafka (`morent.users`).
3. **user-system** потребляет событие и создаёт/обновляет зеркального пользователя (компания `Morent` по умолчанию).

## Безопасность

- В Kafka **никогда не передаётся plaintext-пароль** — только `password_hash` (bcrypt).
- События обёрнуты в `Envelope` с `schema_version`, `source=morent-backend`, уникальным `event_id`.
- **Идемпотентность**: таблица `processed_events` в user-system.
- Ключ сообщения — `email` (нормализованный).

## Топик и типы событий

| Тип | Описание |
|-----|----------|
| `morent.user.registered` | Новый пользователь |
| `morent.user.profile_updated` | Имя, должность, активность |
| `morent.user.password_changed` | Новый bcrypt-хеш |
| `morent.user.deactivated` | Деактивация |

Контракты: `shared/morent-events/`.

## Kafka UI

**http://localhost:8090** → Topics → `morent.users` → Messages (Fetch / Live Mode).

Consumers → группа `user-system-morent-sync`.

## Запуск

```bash
# 1. Kafka + UI
cd Morent-Architecture
docker compose -f infra/kafka/docker-compose.yml up -d

# 2. user-system (consumer)
cd user-system-develop
docker compose up -d --build

# 3. Morent (producer)
cd Morent-project
docker compose up -d --build
```

## Переменные окружения

**Morent** (`KAFKA_ENABLED=true`, `KAFKA_BROKERS`, `MORENT_COMPANY_NAME`) — `Morent-project/.env.example`.

**user-system** (`KAFKA_*`) — `user-system-develop/.env.example`.

Отключить синхронизацию: `KAFKA_ENABLED=false` на нужном сервисе (события не публикуются / не читаются).

## Проверка

1. Регистрация на http://localhost:5173/sign-up
2. Kafka UI: сообщение в `morent.users`
3. `docker logs user-system-api -f` — обработка consumer
4. Пользователь в БД user-system с `morent_user_id`
