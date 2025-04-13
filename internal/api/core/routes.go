package core

import (
	"fmt"
	"strings"

	"github.com/edulinq/autograder/internal/util"
)

var CURRENT_PREFIX string = fmt.Sprintf("/api/v%02d", util.MustGetAPIVersion())

// API routes will be empty until RunServer() is called.
// This is a cached version of api.GetRoutes() for packages that cannot access `internal/api` due to import cycles.
var apiRoutes []Route

// Get an endpoint using the current prefix.
func MakeFullAPIPath(suffix string) string {
	if strings.HasPrefix(suffix, "/") {
		suffix = strings.TrimPrefix(suffix, "/")
	}

	return CURRENT_PREFIX + "/" + suffix
}

func SetAPIRoutes(routes []Route) {
	apiRoutes = routes
}

func GetAPIRoutes() []Route {
	return apiRoutes
}
