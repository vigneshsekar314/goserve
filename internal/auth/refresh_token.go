package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() (string, error) {
	randBytes := make([]byte, 32)
	rand.Read(randBytes)
	refresh_token := hex.EncodeToString(randBytes)
	return refresh_token, nil
}
