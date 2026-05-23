# Observability (Elasticsearch + Prometheus + Grafana)

Централизованные логи, проверки доступности и дашборды для сервисов Morent.

## Компоненты

| Сервис | Порт | Назначение |
|--------|------|------------|
| Elasticsearch | 9200 | Хранение логов и heartbeat |
| Kibana | 5601 | Поиск и дашборды по логам |
| Filebeat | — | Сбор stdout контейнеров Docker |
| Heartbeat | — | HTTP-проверки `/health` и `/ready` |
| Prometheus | 9090 | Метрики blackbox, cAdvisor |
| Grafana | 3001 | Дашборды (Prometheus + Elasticsearch) |

## Запуск

Из корня репозитория:

```bash
make obs-up
```

Перед Filebeat должны быть запущены контейнеры с метками `co.elastic.logs/enabled=true` (Morent backend, payment-service, EmailService).

## Индексы

- `morent-logs-{service}-{log.type}-{date}` — application/http/integration логи
- `morent-heartbeat-{monitor.id}-{date}` — результаты Heartbeat

Поля в логах Go-сервисов (JSON stdout, после Filebeat — префикс `app.`):

| Поле | Описание |
|------|----------|
| `service.name` | Имя сервиса из Docker-метки |
| `app.log_type` | `app`, `http`, `heartbeat`, `integration` |
| `app.path`, `app.method`, `app.status` | HTTP access-log |
| `app.duration_ms` | Время ответа |
| `app.event_type` | Тип Kafka-события |
| `app.template` | Шаблон письма |

## Grafana

URL: http://localhost:3001 (логин `admin` / `admin` по умолчанию).

Дашборды подхватываются автоматически из `grafana/dashboards/`:

| Дашборд | Источник | Содержание |
|---------|----------|------------|
| **Morent Overview** | Prometheus | Health/ready probes, CPU/RAM/сеть контейнеров |
| **Morent All Services** | Prometheus + ES | Сводка по всем 7 app-сервисам + ES/Kibana/Kafka UI |
| **Morent Service Detail** | Prometheus + ES | Детализация по выбранному сервису (переменная `$service`) |
| **Morent Domain APIs** | Elasticsearch | API по доменам: backend, payment, user-system, aggregator, generator, email |
| **Morent HTTP Analytics** | Elasticsearch | RPS, latency, статусы, топ путей, ошибки |
| **Morent User Flows** | Elasticsearch | Пользовательские сценарии (auth, аренда, платежи, …) |
| **Morent Business Events** | Elasticsearch | Kafka, email, напоминания, интеграции, ошибки |
| **Morent Uptime** | ES Heartbeat + Prometheus | Heartbeat-мониторы и blackbox probes |

### Покрытие сервисов

| Сервис | Prometheus `/health` | Filebeat | Heartbeat |
|--------|---------------------|----------|-----------|
| morent-backend | :1488 | да | да |
| payment-service | :8081 | да | да (+ `/ready`) |
| user-system | :8082 | да | да |
| car-aggregator | :8080 | да | да |
| generator-service | :8083 | да | да |
| emailservice-api | :5112 | да | да |
| emailservice-worker | — | да | — |
| elasticsearch, kibana, kafka-ui | внутри obs-сети / :8090 | — | kafka-ui |

Для панелей по логам нужны данные в Elasticsearch (запущенные сервисы + Filebeat). Prometheus-дашборды работают сразу после `make obs-up`.

## Kibana

1. Откройте http://localhost:5601  
2. **Stack Management → Index Patterns**  
3. Pattern `morent-logs-*`, time field `@timestamp`  
4. Отдельно `morent-heartbeat-*` для uptime

## Heartbeat

Мониторы в `heartbeat.yml` бьют в `host.docker.internal` (порты Morent, payment, Email API). На Linux при необходимости замените хост на IP хост-машины.

## Prometheus

`prometheus.yml` — blackbox probes `/health` и `/ready` с меткой `service` для группировки в Grafana.
