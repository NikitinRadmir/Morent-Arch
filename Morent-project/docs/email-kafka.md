# Email-уведомления Morent ↔ EmailService (Kafka)

## Схема

```
Morent backend (создание аренды)
  → Kafka `morent.emails`
  → EmailService.Worker (consumer)
  → внутренняя очередь → рендер шаблона → SendGrid
```

Morent **не вызывает** EmailService по HTTP при бронировании — только публикует событие.

## Топик и события

| Топик | Назначение |
|-------|------------|
| `morent.emails` | Запросы на отправку писем |

| Тип события | Когда |
|-------------|--------|
| `morent.email.send` | `welcome_registered`, `login_notification`, `booking_confirmation`, … |

Ключ сообщения: нормализованный `email` получателя.

Контракты: `shared/morent-events/email.go`, обёртка `Envelope` как для `morent.users`.

## Шаблоны

- `welcome_registered` — после `POST /auth/register`
- `login_notification` — после `POST /auth/login`
- `booking_confirmation` — после успешного `POST /rentals`
- `payment_failed`, `reminder_24h` — зарезервированы в контракте (публикация из Morent по мере появления сценариев)

Файлы: `EmailService/Templates/*.html`

## Запуск

1. Kafka: `docker compose -f infra/kafka/docker-compose.yml up -d`
2. EmailService Worker с `Kafka:Enabled=true` и брокером
3. Morent backend с `KAFKA_ENABLED=true`

## Переменные

**Morent backend:** `KAFKA_TOPIC_EMAILS` (по умолчанию `morent.emails`)

**EmailService.Worker** (`appsettings` / env):

- `Kafka__Enabled=true`
- `Kafka__Brokers=localhost:9092`
- `Kafka__Topic=morent.emails`
- `Kafka__GroupId=email-service`

Если Kafka отключён, письма по-прежнему можно отправить через HTTP `POST /api/v1/emails`.

## Проверка

1. Создать аренду в Morent (авторизованный пользователь)
2. Kafka UI → топик `morent.emails` → сообщение `morent.email.send`
3. Логи Worker: `queued from kafka`
4. `GET /api/v1/emails/{correlationId}` — статус доставки
