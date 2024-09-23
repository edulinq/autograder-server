package users

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []*core.Route = []*core.Route{
	core.NewAPIRoute(core.NewEndpoint(`courses/users/drop`), HandleDrop),
	core.NewAPIRoute(core.NewEndpoint(`courses/users/get`), HandleGet),
}

func GetRoutes() *[]*core.Route {
	return &routes
}
