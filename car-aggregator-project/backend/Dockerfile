FROM golang:1.25 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/car_aggregator .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/car_aggregator usr/local/bin/car_aggregator
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/car_aggregator"]