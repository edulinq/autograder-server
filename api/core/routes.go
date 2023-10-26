package core

import (
    "fmt"
    "strings"
)

const API_VERSION int = 2;
var CURRENT_PREFIX string = fmt.Sprintf("/api/v%02d", API_VERSION);

// Get an endpoint using the current prefix.
func NewEndpoint(suffix string) string {
    if (strings.HasPrefix(suffix, "/")) {
        suffix = strings.TrimPrefix(suffix, "/");
    }

    return CURRENT_PREFIX + "/" + suffix;
}
