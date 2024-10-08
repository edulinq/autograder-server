package users

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []*core.Route = []*core.Route{
	core.NewAPIRoute(`courses/users/drop`, HandleDrop),
	core.NewAPIRoute(`courses/users/enroll`, HandleEnroll),
	core.NewAPIRoute(`courses/users/get`, HandleGet),
	core.NewAPIRoute(`courses/users/list`, HandleList),
}

func GetRoutes() *[]*core.Route {
	return &routes
}
