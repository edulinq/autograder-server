package core

import (
    "strings"
)

const CURRENT_PREFIX string = "/api/v02"

// Get an endpoint using the current prefix.
func NewEndpoint(suffix string) string {
    if (strings.HasPrefix(suffix, "/")) {
        suffix = strings.TrimPrefix(suffix, "/");
    }

    return CURRENT_PREFIX + "/" + suffix;
}
