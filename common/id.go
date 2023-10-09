package common

import (
	"fmt"
    "regexp"
    "strings"
)

// Return a cleaned ID, or an error if the ID cannot be cleaned.
func ValidateID(id string) (string, error) {
    id = strings.ToLower(id);

    if (!regexp.MustCompile(`^[a-z0-9\._\-]+$`).MatchString(id)) {
        return "", fmt.Errorf("IDs must only have letters, digits, and single sequences of periods, underscores, and hyphens, found '%s'.", id);
    }

    if (regexp.MustCompile(`(^[\._\-])|(^[\._\-])$`).MatchString(id)) {
        return "", fmt.Errorf("IDs cannot start or end with periods, underscores, or hyphens, found '%s'.", id);
    }

    return id, nil;
}
