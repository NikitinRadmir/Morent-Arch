FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o backend_server ./cmd/server

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/backend_server .
EXPOSE 1488
CMD ["./backend_server"]


