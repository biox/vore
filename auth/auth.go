package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateSessionToken() string {
	// 32 bytes == 256 bits (AES security margin is 128 bits)
	return generateSecureToken(32)
}

func generateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
