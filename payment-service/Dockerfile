FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY payment-service/go.mod payment-service/go.sum ./payment-service/
COPY shared ./shared
WORKDIR /app/payment-service
RUN go mod download
COPY payment-service ./
RUN CGO_ENABLED=0 GOOS=linux go build -o payment-service ./cmd/server

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/payment-service .
EXPOSE 8081
CMD ["./payment-service"]
