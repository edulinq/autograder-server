package util

import (
   "encoding/json"
   "fmt"
   "io"
   "os"
)

const DEFAULT_PREFIX = "";
const DEFAULT_INDENT = "    ";

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

func JSONFromString(data string, target any) error {
    err := json.Unmarshal([]byte(data), target);
    if (err != nil) {
        return fmt.Errorf("Could not unmarshal JSON string (%s): '%w'.", data, err);
    }

    return nil;
}

func JSONMapFromString(data string) (map[string]any, error) {
    target := make(map[string]any);

    err := json.Unmarshal([]byte(data), &target);
    if (err != nil) {
        return nil, fmt.Errorf("Could not unmarshal JSON map from string(%s): '%w'.", data, err);
    }

    return target, nil;
}

func ToJSON(data any) (string, error) {
    return ToJSONIndentCustom(data, "", "");
}

func ToJSONIndent(data any) (string, error) {
    return ToJSONIndentCustom(data, DEFAULT_PREFIX, DEFAULT_INDENT);
}

func ToJSONIndentCustom(data any, prefix string, indent string) (string, error) {
    bytes, err := json.MarshalIndent(data, prefix, indent);
    if (err != nil) {
        return "", fmt.Errorf("Could not marshal object ('%v'): '%w'.", data, err);
    }

    return string(bytes), nil;
}

func ToJSONFile(data any, path string) error {
    return ToJSONFileIndentCustom(data, path, "", "");
}

func ToJSONFileIndent(data any, path string) error {
    return ToJSONFileIndentCustom(data, path, DEFAULT_PREFIX, DEFAULT_INDENT);
}

func ToJSONFileIndentCustom(data any, path string, prefix string, indent string) error {
    text, err := ToJSONIndentCustom(data, prefix, indent);
    if (err != nil) {
        return err;
    }

    return WriteFile(text, path);
}
