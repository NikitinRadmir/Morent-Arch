# Банк Morent ↔ payment-service (Kafka)

## Схема

```
React (/bank/*) → Morent backend (/bank/*)
  → Kafka `morent.bank.commands`
  → payment-service (обработка + счета)
  → Kafka `morent.bank.responses`
  → Morent backend (request-reply, ~15 с)
  → HTTP ответ клиенту
```

Morent **не хранит** банковские сессии и балансы — только проксирует команды и cookie `morent_bank_session`.

## Топики

| Топик | Назначение |
|-------|------------|
| `morent.bank.commands` | Команды от Morent backend |
| `morent.bank.responses` | Ответы от payment-service |

Ключ сообщения: `request_id` (UUID).

## Команды

`bank.register`, `bank.login`, `bank.logout`, `bank.profile`, `bank.deposit`, `bank.transfer`, `bank.transactions`

Схемы: `shared/morent-events/bank.go`.

## Запуск

1. Kafka: `docker compose -f infra/kafka/docker-compose.yml up -d`
2. payment-service: `docker compose -f payment-service/docker-compose.yml up -d`  
   (или локально с `KAFKA_BROKERS=localhost:9092`)
3. Morent backend с `KAFKA_ENABLED=true` и брокером (см. `.env.example`)

## Переменные

**Morent backend:** `KAFKA_TOPIC_BANK_COMMANDS`, `KAFKA_TOPIC_BANK_RESPONSES`, `KAFKA_GROUP_MORENT_BANK`

**payment-service:** `KAFKA_TOPIC_BANK_COMMANDS`, `KAFKA_TOPIC_BANK_RESPONSES`, `KAFKA_GROUP_PAYMENT_BANK`

Если Kafka недоступен, API банка отвечает **503** (`bank is not available...`).
