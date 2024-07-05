package util

import (
	"crypto/rand"
	"encoding/hex"
	"math"
)

func RandBytes(length int) ([]byte, error) {
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

func RandHex(length int) (string, error) {
	numBytes := int(math.Ceil(float64(length) / 2.0))
	bytes, err := RandBytes(numBytes)
	if err != nil {
		return "", err
	}

	text := hex.EncodeToString(bytes)
	return text[:length], nil
}
