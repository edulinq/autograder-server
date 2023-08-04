package util

import (
   "encoding/json"
   "fmt"
   "io"
   "os"
)

// The target must be a pointer.
func JSONFromFile(path string, target any) error {
    file, err := os.Open(path);
    if (err != nil) {
        return fmt.Errorf("Could not open JSON file (%s): '%w'.", path, err);
    }
    defer file.Close();

    data, err := io.ReadAll(file);
    if (err != nil) {
        return fmt.Errorf("Could not read JSON file (%s): '%w'.", path, err);
    }

    err = json.Unmarshal(data, target);
    if (err != nil) {
        return fmt.Errorf("Could not unmarshal JSON file (%s): '%w'.", path, err);
    }

    return nil;
}

func ToJSON(data any) (string, error) {
    bytes, err := json.Marshal(data);
    if (err != nil) {
        return "", fmt.Errorf("Could not marshal object ('%v'): '%w'.", data, err);
    }

    return string(bytes), nil;
}
