# Автомобильный агрегатор

Проект подразумевает работу с энд-поинтами дилеров для
поиска оптимальных цен с улучшенной системой поиска.

---

## Новые возможности поиска

### Улучшенная система поиска

Агрегатор теперь включает продвинутую систему поиска с:

- **Множественные стратегии поиска**: точный поиск, нормализация через DaData, нечеткий поиск
- **Отказоустойчивость**: автоматические fallback механизмы при сбоях API
- **Ранжирование результатов**: интеллектуальная сортировка по релевантности
- **Поддержка русского языка**: автоматическое распознавание и перевод запросов
- **Подробное логирование**: отслеживание всех операций поиска для отладки

### Стратегии поиска

1. **Точный поиск (Exact Match)**: для запросов с четко указанной маркой и моделью
2. **DaData стратегия**: нормализация запросов через DaData API
3. **Нечеткий поиск (Fuzzy Match)**: для частичных совпадений и опечаток
4. **Fallback механизмы**: резервные методы поиска при сбоях

---

## Запуск проекта

Для запуска проекта Вам понадобятся API ключи от сервисов, используемых
в проекте:

- [dadata.ru](https://dadata.ru/?authorization_popup=1&next=/profile/%23info).
Достаточно будет зарегистрироваться на сайте и подтвердить свою почту для
работы с API. Все необходимые данные можно будет найти в 
[профиле](https://dadata.ru/profile/#info).

- [carapi.app](https://carapi.app/register).
Что касается данного ресурса, то потребуется зайти во вкладку 
["API Credentials"](https://carapi.app/profile/users/api), скопировать 
API Token, а также сгенерировать секрет с помощью кнопки "Generate Secret".

Соответственно все данные необходимо будет скопировать в файл `.env` в 
корневой папке проекта, который будет выглядеть следующим образом:

```env
# API
DADATA_API_KEY=<copy here your "API-ключ">
DADATA_SECRET_KEY=<copy here your "Секретный ключ">
CARAPI_TOKEN=<copy here your "API Token">
CARAPI_SECRET=<copy here your "Api Secret">

# Search Configuration (optional)
SEARCH_CONFIG_FILE=search_config.json
SEARCH_LOG_LEVEL=info
DADATA_TIMEOUT=10s
CARAPI_TIMEOUT=15s
```

---

## Конфигурация поиска

Создайте файл `search_config.json` для настройки параметров поиска:

```json
{
  "dadata": {
    "timeout": "10s",
    "max_results": 5,
    "enabled": true
  },
  "carapi": {
    "timeout": "15s",
    "max_retries": 3,
    "rate_limit": 100
  },
  "fallback": {
    "enabled": true,
    "fuzzy_threshold": 0.7,
    "popular_models": [
      "Golf", "Camry", "Civic", "Corolla", "Focus"
    ]
  },
  "logging": {
    "level": "info",
    "enable_debug_info": false
  }
}
```

---

## Взаимодействие

### Основной эндпоинт поиска

Сервис имеет улучшенный энд-поинт (`http://localhost:8080/search/trims`), 
предоставляющий расширенную информацию об автомобилях.

#### Базовый запрос:
```bash
POST http://localhost:8080/search/trims
Content-Type: application/json

{
  "q": "Volkswagen Golf"
}
```

#### Расширенный запрос с опциями:
```bash
POST http://localhost:8080/search/trims
Content-Type: application/json

{
  "q": "Гольф",
  "filters": {
    "year_after": 2015,
    "year_before": 2023,
    "price_from": 20000,
    "price_to": 50000
  }
}
```

### Примеры запросов

#### 1. Поиск на английском:
```json
{
  "q": "Toyota Camry"
}
```

#### 2. Поиск на русском:
```json
{
  "q": "Тойота Камри"
}
```

#### 3. Поиск с годом:
```json
{
  "q": "2020 BMW X5"
}
```

#### 4. Поиск только модели:
```json
{
  "q": "Golf"
}
```

#### 5. Нечеткий поиск (с опечатками):
```json
{
  "q": "Volswagen Golff"
}
```

### Формат ответа

Улучшенный ответ включает:

```json
{
  "query": "Volkswagen Golf",
  "count": 15,
  "cars": [
    {
      "id": 1,
      "year": 2021,
      "make": "Volkswagen",
      "model": "Golf",
      "trim": "TSI S",
      "description": "TSI S 4dr Hatchback (1.4L 4cyl Turbo 8A)",
      "msrp": 23195,
      "transmission": "Automatic",
      "seats": 5,
      "fuel": 6.7,
      "imageUrl": "https://example.com/image.jpg"
    }
  ],
  "suggestions": [
    "Volkswagen Passat",
    "Volkswagen Jetta"
  ]
}
```

--
```