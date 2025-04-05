package proxy

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var baseRoutes []core.Route = []core.Route{
	core.MustNewAPIRoute(`courses/assignments/submissions/proxy/resubmit`, HandleResubmit),
	core.MustNewAPIRoute(`courses/assignments/submissions/proxy/submit`, HandleSubmit),
}

func GetRoutes() *[]core.Route {
	routes := make([]core.Route, 0)

	routes = append(routes, baseRoutes...)

	return &routes
}
