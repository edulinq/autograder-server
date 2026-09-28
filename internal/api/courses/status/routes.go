package status

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.MustNewAPIRoute(`courses/status/remove`, HandleRemove),
	core.MustNewAPIRoute(`courses/status/set`, HandleSet),
	core.MustNewAPIRoute(`courses/status/list`, HandleList),
}

func GetRoutes() *[]core.Route {
	return &routes
}
