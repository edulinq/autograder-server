package util

import (
    "crypto/rand"
)

func RandBytes(length int) ([]byte, error) {
    buffer := make([]byte, length);
    _, err := rand.Read(buffer);
    if (err != nil) {
        return nil, err;
    }

    return buffer, nil;
}
