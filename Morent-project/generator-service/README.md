# Generator Microservice

Микросервис для генерации паролей и QR-кодов.

## Возможности

- Автоматическая генерация паролей (10-14 символов)
- Генерация паролей по маске
- Генерация QR-кодов

## API Endpoints

### Password Generation

#### Автоматическая генерация
```
GET /api/v1/password
```

Ответ:
```json
{
  "password": "aB3$xY9!mN2",
  "mask": "lLdsLdslLd",
  "length": 11
}
```

#### Генерация по маске
```
GET /api/v1/password/mask?mask=LLLdddss
```

Символы маски:
- `l` = lowercase (a-z)
- `L` = uppercase (A-Z)
- `d` = digit (0-9)
- `s` = special (!@#$%^&*()-_=+[]{}?)

Ответ:
```json
{
  "password": "ABC123!@",
  "mask": "LLLdddss"
}
```

### QR Code Generation

```
GET /api/v1/qrcode?data=HelloWorld&size=256
```

Параметры:
- `data` (обязательный): данные для кодирования
- `size` (опциональный): размер в пикселях (по умолчанию 256, минимум 64)

Ответ: PNG изображение

## Запуск

### Локально
```bash
cd generator-service
go mod download
go run cmd/server/main.go
```

Или используя Makefile:
```bash
make run
```

### Docker
```bash
cd generator-service
docker build -t generator-service .
docker run -p 8080:8080 generator-service
```

Или используя docker-compose:
```bash
docker-compose up -d
```

## Примеры использования

```bash
# Генерация пароля
curl http://localhost:8080/api/v1/password

# Генерация по маске
curl "http://localhost:8080/api/v1/password/mask?mask=LLLdddss"

# Генерация QR-кода
curl "http://localhost:8080/api/v1/qrcode?data=HelloWorld&size=256" --output qr.png
```

## Архитектура

```
generator-service/
├── cmd/
│   └── server/
│       └── main.go          # Точка входа
├── internal/
│   ├── generators/          # Логика генерации
│   │   ├── password.go
│   │   └── qrcode.go
│   └── handlers/            # HTTP handlers
│       └── generator_handler.go
├── Dockerfile
├── go.mod
└── README.md
```

## Зависимости

- `github.com/gin-gonic/gin` - HTTP framework
- `github.com/skip2/go-qrcode` - QR code generation
