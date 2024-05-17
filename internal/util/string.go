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
