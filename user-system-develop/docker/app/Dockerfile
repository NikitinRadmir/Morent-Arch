FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копируем весь код
COPY . .

# Скачиваем зависимости и создаем go.sum
RUN go mod download
RUN go mod tidy

# Собираем приложение
RUN go build -o main ./app

EXPOSE 8080

CMD ["./main"]
