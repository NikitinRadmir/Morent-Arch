# Car Aggregator (Simplified)

Упрощенный Car Aggregator API - простой прокси-сервис для получения данных об автомобилях из внешних API.

## Описание

Car Aggregator - это легковесный HTTP сервис, который интегрируется с внешними API для поиска информации об автомобилях:
- **dadata.ru** - для нормализации названий автомобилей
- **carapi.app** - для получения данных об автомобилях и комплектациях

Сервис предназначен для интеграции с платформой Morent и предоставляет простой REST API.

## Запуск проекта

### Требования
- Go 1.21+
- Docker и Docker Compose (опционально)

### Настройка API ключей

1. **dadata.ru**
   - Зарегистрируйтесь на [dadata.ru](https://dadata.ru/?authorization_popup=1&next=/profile/%23info)
   - Подтвердите email
   - Скопируйте API-ключ и Секретный ключ из [профиля](https://dadata.ru/profile/#info)

2. **carapi.app**
   - Зарегистрируйтесь на [carapi.app](https://carapi.app/register)
   - Перейдите в ["API Credentials"](https://carapi.app/profile/users/api)
   - Скопируйте API Token и сгенерируйте Api Secret

3. **Конфигурация**
   - Скопируйте `.env.example` в `.env`
   - Заполните API ключи в файле `.env`

### Локальный запуск

```bash
# Установка зависимостей
go mod download

# Запуск сервиса
go run main.go
```

### Запуск через Docker

```bash
# Сборка и запуск
docker-compose up --build

# Запуск в фоне
docker-compose up -d --build
```

## API Endpoints

### GET /search
Поиск автомобилей по названию.

**Запрос:**
```json
{
  "q": "Volkswagen Golf"
}
```

**Ответ:**
```json
{
  "info": "Vehicle Information:\n\nAvailable Trims (Total: 39):\n- 2015 Volkswagen Golf TSI Launch Edition (MSRP: $17995)\n TSI Launch Edition 2dr Hatchback (1.8L 4cyl Turbo 5M)\n..."
}
```

### GET /search/trims?car_id=123
Получение комплектаций для конкретного автомобиля.

**Ответ:**
```json
{
  "info": "Available Trims (Total: 5):\n- TSI S (MSRP: $19295)\n TSI S 2dr Hatchback (1.8L 4cyl Turbo 5M)\n..."
}
```

### GET /health
Проверка состояния сервиса.

**Ответ:**
```json
{
  "status": "ok",
  "service": "car-aggregator"
}
```

## Архитектура

Упрощенная архитектура без базы данных и очередей:

```
Morent → Car Aggregator → External APIs (dadata.ru, carapi.app)
```

### Структура проекта

```
car-aggregator-simplified/
├── main.go                 # Точка входа приложения
├── internal/
│   ├── handlers/           # HTTP обработчики
│   │   ├── router.go       # Настройка маршрутов
│   │   └── search.go       # Обработчик поиска
│   └── services/           # Сервисы для внешних API
│       ├── dadata.go       # Клиент dadata.ru
│       └── carapi.go       # Клиент carapi.app
├── go.mod                  # Go модули
├── Dockerfile              # Docker образ
├── docker-compose.yml      # Docker Compose конфигурация
├── .env.example            # Пример переменных окружения
└── README.md               # Документация
```

## Интеграция с Morent

Сервис полностью совместим с существующей интеграцией Morent:
- Использует те же эндпоинты `/search` и `/search/trims`
- Возвращает данные в том же формате
- Работает на том же порту `:8080`

Настройка в Morent:
```env
AGGREGATOR_BASE_URL=http://localhost:8080
```

## Мониторинг

- **Health Check**: `GET /health` - проверка состояния сервиса
- **Логи**: все ошибки логируются в stdout
- **Метрики**: базовые HTTP метрики через Gin framework

## Разработка

### Тестирование
```bash
# Запуск тестов
go test ./...

# Тесты с покрытием
go test -cover ./...
```

### Линтинг
```bash
# Установка golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Запуск линтера
golangci-lint run
```