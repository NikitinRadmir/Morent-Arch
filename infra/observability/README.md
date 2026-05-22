# Observability (Elasticsearch)

Централизованные логи и проверки доступности сервисов Morent.

## Компоненты

| Сервис | Порт | Назначение |
|--------|------|------------|
| Elasticsearch | 9200 | Хранение логов и heartbeat |
| Kibana | 5601 | Поиск и дашборды |
| Filebeat | — | Сбор stdout контейнеров Docker |
| Heartbeat | — | HTTP-проверки `/health` и `/ready` |

## Запуск

Из корня репозитория:

```bash
make obs-up
```

Перед Filebeat должны быть запущены контейнеры с метками `co.elastic.logs/enabled=true` (Morent backend, payment-service, EmailService).

## Индексы

- `morent-logs-{service}-{log.type}-{date}` — application/http/integration логи
- `morent-heartbeat-{monitor.id}-{date}` — результаты Heartbeat

Поля в логах Go-сервисов (JSON stdout):

- `service` — имя сервиса (`morent-backend`, `payment-service`, …)
- `log_type` — `app` | `http` | `heartbeat` | `integration` | `health`

## Kibana

1. Откройте http://localhost:5601  
2. **Stack Management → Index Patterns**  
3. Pattern `morent-logs-*`, time field `@timestamp`  
4. Отдельно `morent-heartbeat-*` для uptime

## Heartbeat

Мониторы в `heartbeat.yml` бьют в `host.docker.internal` (порты Morent, payment, Email API). На Linux при необходимости замените хост на IP хост-машины.
