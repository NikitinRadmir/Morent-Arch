package generators

import (
	"strconv"

	"github.com/skip2/go-qrcode"
)

// GenerateQR генерирует QR-код из данных
func GenerateQR(data string, size int) ([]byte, error) {
	return qrcode.Encode(data, qrcode.Medium, size)
}

// ParseSize парсит размер QR-кода из строки
func ParseSize(val string) int {
	if val == "" {
		return 256
	}
	num, err := strconv.Atoi(val)
	if err != nil || num < 64 {
		return 256
	}
	return num
}
