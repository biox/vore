package auth

import (
	"fmt"
	"time"
)

func GenerateSessionToken() string {
	// TODO: don't use the time ffs
	token := fmt.Sprintf("%x", time.Now().UnixNano())
	return token
}
