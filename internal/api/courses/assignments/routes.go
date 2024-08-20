package assignments

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/assignments/submissions"
)

var routes []*core.Route = []*core.Route{
	core.NewAPIRoute(core.NewEndpoint(`courses/assignments/get`), HandleGet),
	core.NewAPIRoute(core.NewEndpoint(`courses/assignments/list`), HandleList),
}

func GetRoutes() *[]*core.Route {
	routes = append(routes, *(submissions.GetRoutes())...)

	return &routes
}
