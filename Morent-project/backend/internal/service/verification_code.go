package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

const emailVerificationTTL = 15 * time.Minute

func generateEmailVerificationCode() (string, time.Time, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", time.Time{}, err
	}
	return fmt.Sprintf("%06d", n.Int64()), time.Now().UTC().Add(emailVerificationTTL), nil
}
