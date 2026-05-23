# Demo runbook: TC/UC защита Morent

Сценарий построен так: use cases показываем руками на фронте, test cases подтверждаем API-скриптом, отказ микросервиса показываем отдельно через controlled outage.

## 0. Если Docker уже запущен

Полный `make up` не нужен. На Windows `make` может быть не установлен, поэтому используем PowerShell и `docker compose`.

Быстрая проверка:

```powershell
Invoke-RestMethod "http://localhost:1488/health"
Invoke-RestMethod "http://localhost:8082/health"
Invoke-RestMethod "http://localhost:8080/health"
Invoke-RestMethod "http://localhost:9090/-/ready"
Invoke-RestMethod "http://localhost:3001/api/health"
```

## 1. Use Cases показываем руками на фронте

Открыть Morent UI:

```text
http://localhost:5173
```

Показать руками:

1. UC-01: регистрация, подтверждение email, вход.
2. UC-02: каталог, выбор авто, выбор дат, оформление аренды.
3. UC-03: добавление автомобиля в избранное.
4. UC-04: админка, поиск в агрегаторе, импорт авто.
5. UC-06: комментарий после аренды.
6. UC-07: админ-панель и CRUD/просмотр данных.

Доступ администратора:

```text
admin@morent.com
admin123
```

Объяснение для преподавателя:

```text
Use cases я показываю через интерфейс, потому что это пользовательские сценарии.
Test cases я подтверждаю через API, потому что там важны статусы, тело ответа и негативные проверки.
```

## 2. Test Cases прогоняем одной командой

Из корня проекта:

```powershell
node .\scripts\demo-use-cases.mjs
```

Скрипт проверяет:

- красивая публичная ошибка логина вместо Go validator dump;
- UC-01: регистрация, код из локальной БД, подтверждение email, вход;
- UC-02 / TC-03: каталог, банковская сессия, пополнение, успешная аренда `201`;
- TC-04: повторная аренда на тот же период возвращает `409`;
- UC-03: избранное не создает дубль;
- UC-06: комментарий разрешен после аренды;
- TC-23 / UC-07: обычный пользователь получает `403` на admin endpoint;
- TC-08 / UC-04: админ ищет авто в агрегаторе и импортирует его;
- UC-08: user-system создает роль в компании.

Ожидаемый итог: все строки начинаются с `PASS`.

## 3. TC-09: роняем агрегатор и получаем 503

Остановить car-aggregator:

```powershell
docker compose --project-directory E:\Morent-Arch\car-aggregator-project -f E:\Morent-Arch\car-aggregator-project\docker-compose.yml down
```

Залогиниться админом в PowerShell-сессию:

```powershell
$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession

Invoke-RestMethod `
  -Uri "http://localhost:1488/auth/login" `
  -Method POST `
  -ContentType "application/json" `
  -Body (@{ email = "admin@morent.com"; password = "admin123" } | ConvertTo-Json -Compress) `
  -WebSession $session
```

Дернуть admin endpoint агрегатора:

```powershell
try {
  Invoke-WebRequest `
    -Uri "http://localhost:1488/Admin/Aggregator/Cars?q=toyota" `
    -WebSession $session `
    -UseBasicParsing
} catch {
  [int]$_.Exception.Response.StatusCode
}
```

Ожидаемо:

```text
503
```

Показать, что основной Morent backend жив:

```powershell
Invoke-RestMethod "http://localhost:1488/health"
```

Ожидаемо:

```json
{"status":"ok","service":"morent-backend"}
```

Вернуть агрегатор:

```powershell
docker compose --project-directory E:\Morent-Arch\car-aggregator-project -f E:\Morent-Arch\car-aggregator-project\docker-compose.yml up -d --build
```

Проверить, что агрегатор снова жив:

```powershell
Invoke-RestMethod "http://localhost:8080/health"
```

## 4. Observability

Kibana:

```text
http://localhost:5601
```

Data view:

```text
morent-logs-*
```

Фильтры:

```text
service.name : "morent-backend"
app.status >= 400
```

Grafana:

```text
http://localhost:3001
admin / admin
```

Prometheus/Grafana запросы:

```promql
probe_success
avg_over_time(probe_success[5m])
up
```

## 5. Короткая речь

```text
Сначала я показываю бизнес-сценарии на фронте. Затем тем же требованиям соответствуют автоматические API-проверки: успешные сценарии, негативные сценарии, роли и интеграции. После этого я показываю отказ агрегатора: сервис агрегатора выключен, Morent возвращает контролируемый 503 и сам остается доступен по health-check.
```
