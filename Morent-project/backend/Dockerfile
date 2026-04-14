FROM golang:1.21-alpine as builder
WORKDIR /app
COPY . .
RUN cd backend && go build -o backend_server ./cmd/server

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/backend/backend_server .
COPY backend/config.json ./config.json
EXPOSE 1488
CMD ["./backend_server"]


