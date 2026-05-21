package bank

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const cardBIN = "4276"

// GenerateVirtualCard создаёт уникальный номер карты, срок MM-YY (+4 года), CVV и держателя.
func GenerateVirtualCard(holder string, registeredAt time.Time, cardExists func(string) bool) (number, expDate, cvv, cardHolder string, err error) {
	holder = strings.TrimSpace(holder)
	if holder == "" {
		holder = "CLIENT"
	}
	cardHolder = strings.ToUpper(holder)

	exp := registeredAt.UTC().AddDate(4, 0, 0)
	expDate = fmt.Sprintf("%02d-%02d", int(exp.Month()), exp.Year()%100)

	cvv, err = randomDigits(3)
	if err != nil {
		return "", "", "", "", err
	}

	for attempt := 0; attempt < 64; attempt++ {
		number, err = randomCardNumber()
		if err != nil {
			return "", "", "", "", err
		}
		if cardExists == nil || !cardExists(number) {
			return number, expDate, cvv, cardHolder, nil
		}
	}
	return "", "", "", "", fmt.Errorf("failed to generate unique card number")
}

func randomCardNumber() (string, error) {
	var body [11]byte
	for i := range body {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		body[i] = byte('0' + n.Int64())
	}
	partial := cardBIN + string(body[:])
	check, err := luhnCheckDigit(partial)
	if err != nil {
		return "", err
	}
	return partial + check, nil
}

func luhnCheckDigit(number string) (string, error) {
	sum := 0
	alt := true
	for i := len(number) - 1; i >= 0; i-- {
		if number[i] < '0' || number[i] > '9' {
			return "", fmt.Errorf("invalid digit")
		}
		n := int(number[i] - '0')
		if alt {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		alt = !alt
	}
	return fmt.Sprintf("%d", (10-(sum%10))%10), nil
}

func randomDigits(n int) (string, error) {
	var b strings.Builder
	for i := 0; i < n; i++ {
		d, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		b.WriteByte(byte('0' + d.Int64()))
	}
	out := b.String()
	if n == 3 && out == "000" {
		return "137", nil
	}
	return out, nil
}

// NormalizeCardNumber оставляет только 16 цифр номера карты.
func NormalizeCardNumber(raw string) (string, error) {
	var b strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	d := b.String()
	if len(d) != 16 {
		return "", fmt.Errorf("invalid card number")
	}
	return d, nil
}

// FormatCardNumber группирует 16 цифр по 4.
func FormatCardNumber(number string) string {
	if len(number) != 16 {
		return number
	}
	return number[0:4] + " " + number[4:8] + " " + number[8:12] + " " + number[12:16]
}
