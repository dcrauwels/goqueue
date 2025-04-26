package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() (string, error) {
	// get the hex
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}

	// hex to string
	encodedKey := hex.EncodeToString(key)
	return encodedKey, nil
}
