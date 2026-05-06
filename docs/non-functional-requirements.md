# Нефункциональные требования

## Покрываемые сервисы

- `Morent core`
- `car-aggregator-project`
- `user-system-develop`
- `payment-service`
- `EmailService`
- `generator-service`

## Список требований

| ID | Сервис | Категория | Требование |
| --- | --- | --- | --- |
| `NFR-01` | Morent core | Availability | Доступность API Morent должна быть не ниже `95%` в месяц на учебном стенде. |
| `NFR-02` | Morent core | Performance | Для ключевых REST-операций Morent p95 latency должна быть <= `400 ms` при нагрузке до `50 RPS`. |
| `NFR-03` | Morent core | Reliability | Система не должна допускать двойную аренду на пересекающиеся интервалы. |
| `NFR-04` | Morent core | Security | Доступ к admin endpoints должен быть ограничен ролью `admin`. |
| `NFR-05` | car-aggregator | Performance | Поисковый запрос агрегатора должен обрабатываться с p95 <= `1500 ms` при нагрузке до `15 RPS`. |
| `NFR-06` | car-aggregator | Reliability | При недоступности DaData/CarAPI должен срабатывать fallback без падения сервиса. |
| `NFR-07` | user-system | Security | JWT-токены должны проверяться на каждом защищенном endpoint. |
| `NFR-08` | user-system | Reliability | Ошибки RBAC должны возвращаться в единообразном формате (`401/403`). |
| `NFR-09` | payment-service | Consistency | Операции перевода должны быть идемпотентны и консистентны при повторных запросах. |
| `NFR-10` | payment-service | Performance | p95 для денежных операций должна быть <= `300 ms` при нагрузке до `80 RPS` в памяти. |
| `NFR-11` | EmailService | Reliability | Очередь отправки должна поддерживать retry при временных ошибках SMTP/API провайдера. |
| `NFR-12` | EmailService | Security | Вебхуки статусов должны проходить проверку HMAC-подписи. |
| `NFR-13` | generator-service | Performance | Генерация пароля должна выполняться <= `100 ms`, QR <= `300 ms` для размера `256x256`. |
| `NFR-14` | generator-service | Security | Сервис не должен логировать чувствительные значения сгенерированных паролей в открытом виде. |
| `NFR-15` | Все сервисы | Deployability | Порты и параметры запуска должны подтягиваться из `.env` и `docker compose` без хардкода. |
| `NFR-16` | Все сервисы | Observability | Для каждого сервиса должны быть health/readiness checks и технические логи. |
| `NFR-17` | Все сервисы | Maintainability | Сервисы должны сохранять слоистую структуру (transport/service/repository/domain). |
| `NFR-18` | Все сервисы | Recoverability | Целевой RPO для БД/состояния — до `24 часов`, RTO — до `30 минут` в учебном окружении. |

## Метрики приемки

| NFR | Метрика / признак приемки |
| --- | --- |
| `NFR-01` | Мониторинг/логи подтверждают uptime >= 95% |
| `NFR-02` | Нагрузочный тест показывает p95 <= 400 ms на базовых endpoint |
| `NFR-05` | Нагрузочный тест aggregator-flow показывает p95 <= 1500 ms |
| `NFR-09` | Повторный запрос с тем же idempotency key не создает дубль операции |
| `NFR-11` | Ошибка провайдера приводит к retry, а не к потере письма |
| `NFR-12` | Невалидная подпись webhook отклоняется |
| `NFR-15` | `docker compose up` поднимает сервисы без ручной правки портов в yaml |
| `NFR-16` | Для сервисов доступны endpoints диагностики |
| `NFR-18` | Backup-процедура выполняется минимум раз в сутки |

## Риски

1. Внешние API (CarAPI, DaData, SendGrid/SMTP) могут давать нестабильные задержки.
2. Учебное окружение обычно имеет ограниченные CPU/RAM, это влияет на p95.
3. При частичных отказах интеграций важно корректно реализовать timeout/retry/circuit-breaker.
4. Для межсервисных flow без единого correlation-id усложняется сквозная диагностика.

