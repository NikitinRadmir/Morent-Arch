package generators

import (
	"crypto/rand"
	"errors"
	"math/big"
)

const (
	lowerLetters   = "abcdefghijklmnopqrstuvwxyz"
	upperLetters   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits         = "0123456789"
	specialSymbols = "!@#$%^&*()-_=+[]{}?"
)

// GeneratePasswordByMask генерирует пароль по заданной маске
// Символы маски: l=lowercase, L=uppercase, d=digit, s=special
func GeneratePasswordByMask(mask string) (string, error) {
	if mask == "" {
		return "", errors.New("mask cannot be empty")
	}

	password := make([]byte, len(mask))
	for i, char := range mask {
		var charset string
		switch char {
		case 'l':
			charset = lowerLetters
		case 'L':
			charset = upperLetters
		case 'd':
			charset = digits
		case 's':
			charset = specialSymbols
		default:
			password[i] = byte(char)
			continue
		}

		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password[i] = charset[idx.Int64()]
	}

	return string(password), nil
}

// GenerateRandomMask создает случайную маску для пароля заданной длины
func GenerateRandomMask(length int) (string, error) {
	if length < 10 || length > 14 {
		return "", errors.New("password length must be between 10 and 14")
	}

	maskChars := []rune{'l', 'L', 'd', 's'}
	mask := make([]rune, length)

	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(maskChars))))
		if err != nil {
			return "", err
		}
		mask[i] = maskChars[idx.Int64()]
	}

	return string(mask), nil
}

// GeneratePasswordAuto автоматически генерирует пароль со случайной длиной 10-14 символов
func GeneratePasswordAuto() (password string, mask string, err error) {
	lengthRange := 5 // 10-14 = 5 вариантов
	lengthIdx, err := rand.Int(rand.Reader, big.NewInt(int64(lengthRange)))
	if err != nil {
		return "", "", err
	}
	length := 10 + int(lengthIdx.Int64())

	mask, err = GenerateRandomMask(length)
	if err != nil {
		return "", "", err
	}

	password, err = GeneratePasswordByMask(mask)
	if err != nil {
		return "", "", err
	}

	return password, mask, nil
}
