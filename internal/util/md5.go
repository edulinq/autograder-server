package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func MD5FileHex(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("Failed to open file '%s' for MD5 hashing: '%w'.", path, err)
	}
	defer file.Close()

	hash := md5.New()

	_, err = io.Copy(hash, file)
	if err != nil {
		return "", fmt.Errorf("Failed to copy file '%s' for MD5 hashing: '%w'.", path, err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func MD5StringHex(content string) (string, error) {
	hash := md5.New()

	_, err := io.WriteString(hash, content)
	if err != nil {
		return "", fmt.Errorf("Failed to copy string contents for MD5 hashing: '%w'.", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
