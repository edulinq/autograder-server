package util

import (
    "encoding/base64"
)

func Base64Encode(data []byte) string {
    return base64.StdEncoding.EncodeToString(data);
}

func Base64Decode(payload string) ([]byte, error) {
    return base64.StdEncoding.DecodeString(payload);
}
