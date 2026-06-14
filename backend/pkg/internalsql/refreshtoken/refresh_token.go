package refreshtoken

import (
	"crypto/rand"
	"encoding/hex"
)

func Generate() (string, error) {
	buffer := make([]byte, 24)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(buffer), nil
}
