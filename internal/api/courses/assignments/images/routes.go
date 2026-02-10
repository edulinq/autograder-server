package images

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var baseRoutes []core.Route = []core.Route{
	core.MustNewAPIRoute(`courses/assignments/images/fetch`, HandleFetch),
}

func GetRoutes() *[]core.Route {
	routes := make([]core.Route, 0)

	routes = append(routes, baseRoutes...)

	return &routes
}
