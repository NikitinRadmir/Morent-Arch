FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o backend_server ./cmd/server

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/backend_server .
ARG BACKEND_INTERNAL_PORT=1488
ARG GRPC_PORT=50051
EXPOSE ${BACKEND_INTERNAL_PORT}
EXPOSE ${GRPC_PORT}
CMD ["./backend_server"]


