package util

import (
    "fmt"
    "strings"
)

func DockerfilePathQuote(path string) string {
    return fmt.Sprintf("\"%s\"", strings.ReplaceAll(path, "\"", "\\\""));
}
