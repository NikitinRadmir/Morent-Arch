FROM golang:1.23-alpine AS builder

WORKDIR /build/user-system-develop

COPY shared/morent-events /build/shared/morent-events
COPY user-system-develop/go.mod user-system-develop/go.sum ./
RUN go mod download

COPY user-system-develop/ .
RUN go build -o main ./app

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /build/user-system-develop/main .
EXPOSE 8080
CMD ["./main", "--clear-db=false"]
