package logs

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []*core.Route = []*core.Route{
	core.NewAPIRoute(core.NewEndpoint(`logs/query`), HandleQuery),
}

func GetRoutes() *[]*core.Route {
	return &routes
}
