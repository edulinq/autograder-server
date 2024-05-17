package util

import (
	"crypto/sha256"
	"encoding/hex"
)

func Sha256Hex(data []byte) string {
	sha := sha256.Sum256(data)
	return hex.EncodeToString(sha[:])
}

func Sha256HexFromString(data string) string {
	return Sha256Hex([]byte(data))
}

// Convert an object to JSON, then hash the JSON.
func Sha256HashFromJSONObject(object any) (string, error) {
	json, err := ToJSON(object)
	if err != nil {
		return "", err
	}

	return Sha256HexFromString(json), nil
}
