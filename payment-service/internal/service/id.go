package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"
)

var fallbackIDSeq uint64

func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%x%08x", time.Now().UnixNano(), atomic.AddUint64(&fallbackIDSeq, 1))
	}
	return hex.EncodeToString(b[:])
}
