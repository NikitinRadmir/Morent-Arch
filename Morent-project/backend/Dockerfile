FROM golang:1.24-alpine AS builder
WORKDIR /build/Morent-project/backend

COPY shared/morent-events /build/shared/morent-events
COPY shared/obslog /build/shared/obslog
COPY Morent-project/backend/go.mod Morent-project/backend/go.sum ./
RUN go mod download

COPY Morent-project/backend/ .
RUN go build -o backend_server ./cmd/server

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /build/Morent-project/backend/backend_server .
EXPOSE 1488
CMD ["./backend_server"]
