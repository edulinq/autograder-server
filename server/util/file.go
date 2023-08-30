package util

import (
    "os"
)

func WriteBinaryFile(data []byte, path string) error {
    return os.WriteFile(path, data, 0644);
}

func WriteFile(text string, path string) error {
    return WriteBinaryFile([]byte(text), path);
}
