# Observability runbook: Elastic, Kibana, Prometheus, Grafana

Этот файл нужен как практическая шпаргалка: где смотреть логи, где смотреть метрики и как быстро понять, какой сервис упал.

## Адреса

| Инструмент | URL | Для чего |
| --- | --- | --- |
| Kibana | http://localhost:5601 | Поиск по логам и событиям |
| Elasticsearch | http://localhost:9200 | API хранилища логов |
| Prometheus | http://localhost:9090 | Сырые метрики и PromQL |
| Grafana | http://localhost:3001 | Дашборды и алерты, логин `admin` / `admin` |
| Kafka UI | http://localhost:8090 | Топики Kafka и сообщения |

## Что где смотреть

Elastic/Kibana отвечает на вопрос: "что именно произошло в приложении?". Там ищем конкретный HTTP-запрос, ошибку, stack/error message, статус, путь, сервис.

Prometheus/Grafana отвечает на вопрос: "жив ли сервис и как он себя чувствует во времени?". Там смотрим uptime, latency, CPU/RAM контейнеров, падение health-check.

## Kibana: первый запуск

1. Открыть http://localhost:5601.
2. Перейти в `Stack Management -> Data Views`.
3. Создать data view `morent-logs-*`, time field `@timestamp`.
4. Создать второй data view `morent-heartbeat-*`, time field `@timestamp`.
5. Перейти в `Discover` и выбрать нужный data view.

Полезные фильтры в Discover:

```text
service.name : "morent-backend"
app.status >= 500
app.path : "/bank/*"
app.log_type : "http"
message : "password"
```

Для проверки доступности сервисов через Heartbeat:

```text
monitor.status : "down"
monitor.name : "Morent Backend"
```

## Elasticsearch API

Список индексов:

```bash
curl http://localhost:9200/_cat/indices?v
```

Последние ошибки 5xx:

```bash
curl -X POST http://localhost:9200/morent-logs-*/_search -H "Content-Type: application/json" -d "{\"size\":10,\"sort\":[{\"@timestamp\":\"desc\"}],\"query\":{\"range\":{\"app.status\":{\"gte\":500}}}}"
```

Логи конкретного сервиса:

```bash
curl -X POST http://localhost:9200/morent-logs-*/_search -H "Content-Type: application/json" -d "{\"size\":20,\"sort\":[{\"@timestamp\":\"desc\"}],\"query\":{\"term\":{\"service.name\":\"morent-backend\"}}}"
```

## Grafana и Prometheus

В Grafana открыть http://localhost:3001, зайти `admin` / `admin`, затем `Dashboards -> Morent Overview`.

Полезные PromQL-запросы:

```promql
probe_success
avg_over_time(probe_success[5m])
probe_duration_seconds
up
sum by (name) (rate(container_cpu_usage_seconds_total[5m]))
sum by (name) (container_memory_working_set_bytes)
```

Как читать:

- `probe_success = 1` значит blackbox health-check прошел.
- `probe_success = 0` значит сервис снаружи не отвечает или health endpoint вернул плохой статус.
- `up = 0` значит Prometheus не может собрать метрики с target.
- Рост `probe_duration_seconds` значит сервис еще жив, но отвечает медленно.

## Быстрый разбор 500/502/503

1. На фронте посмотреть текст ошибки. Теперь сетевые ошибки и 500/502/503 проходят через общий клиент и не должны падать как `Unexpected end of JSON input`.
2. В Kibana открыть `morent-logs-*` и отфильтровать `app.status >= 500`.
3. Добавить фильтр `service.name`, например `morent-backend`, `payment-service`, `email-service`.
4. В Grafana проверить `probe_success` по сервисам.
5. Если `503`, чаще всего зависимый сервис недоступен: bank, generator, Kafka или внешний integration.
6. Если `502`, сервис доступен, но вернул некорректный ответ upstream.
7. Если `500`, это внутренняя ошибка текущего сервиса; нужно смотреть конкретный лог рядом по времени.

## Что алертить первым

- `probe_success == 0` дольше 2 минут для backend/payment/generator/user-system/email.
- Рост 5xx в `morent-backend`.
- Недоступность Prometheus target: `up == 0`.
- Высокую память контейнера через `container_memory_working_set_bytes`.
