module morent-arch/payment-service

go 1.23

replace morent-events => ../shared/morent-events

require (
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/segmentio/kafka-go v0.4.47
	golang.org/x/crypto v0.24.0
	morent-events v0.0.0
)

require (
	github.com/klauspost/compress v1.17.6 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
)
