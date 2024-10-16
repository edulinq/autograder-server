package core

import (
	"fmt"
	"strings"

	"github.com/edulinq/autograder/internal/util"
)

var CURRENT_PREFIX string = fmt.Sprintf("/api/v%02d", util.MustGetAPIVersion())

// Get an endpoint using the current prefix.
func makeFullAPIPath(suffix string) string {
	if strings.HasPrefix(suffix, "/") {
		suffix = strings.TrimPrefix(suffix, "/")
	}

	return CURRENT_PREFIX + "/" + suffix
}
