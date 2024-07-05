package util

import (
	"fmt"
	"strings"
)

func BaseString(obj any) string {
	json, err := ToJSON(obj)
	if err != nil {
		// Explicitly use Go-Syntax (%#v) to avoid loops with overwritten String() methods.
		return fmt.Sprintf("%#v", obj)
	}

	return json
}

func JoinStrings(delim string, parts ...string) string {
	return strings.Join(parts, delim)
}

// A utility function for creating a string pointer for a string literal.
// Should mostly be used for testing.
func StringPointer(text string) *string {
	return &text
}

func PointerToString(target *string) string {
	if target == nil {
		return ""
	}

	return *target
}
