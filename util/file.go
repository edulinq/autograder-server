package util

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

func ReadFile(path string) (string, error) {
    data, err := os.ReadFile(path);
    if (err != nil) {
        return "", err;
    }

    return string(data[:]), nil;
}

func WriteBinaryFile(data []byte, path string) error {
    return os.WriteFile(path, data, 0644);
}

func WriteFile(text string, path string) error {
    return WriteBinaryFile([]byte(text), path);
}

// Read a separated file into a slice of slices.
func ReadSeparatedFile(path string, delim string, skipRows int) ([][]string, error) {
    file, err := os.Open(path);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to open file '%s': '%w'.", path, err);
    }
    defer file.Close();

    rows := make([][]string, 0);

    scanner := bufio.NewScanner(file);
    scanner.Split(bufio.ScanLines);

    for scanner.Scan() {
        if (skipRows > 0) {
            skipRows--;
            continue;
        }

        rows = append(rows, strings.Split(scanner.Text(), delim));
    }

    return rows, nil;
}
