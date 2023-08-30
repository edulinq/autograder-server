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
    return ToJSONIndent(data, "", "");
}

func ToJSONIndent(data any, prefix string, indent string) (string, error) {
    bytes, err := json.MarshalIndent(data, prefix, indent);
    if (err != nil) {
        return "", fmt.Errorf("Could not marshal object ('%v'): '%w'.", data, err);
    }

    return string(bytes), nil;
}

func ToJSONFile(data any, path string) error {
    return ToJSONFileIndent(data, path, "", "");
}

func ToJSONFileIndent(data any, path string, prefix string, indent string) error {
    text, err := ToJSONIndent(data, prefix, indent);
    if (err != nil) {
        return err;
    }

    return WriteFile(text, path);
}
