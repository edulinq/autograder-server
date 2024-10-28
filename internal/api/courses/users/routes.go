package users

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.MustNewAPIRoute(`courses/users/drop`, HandleDrop),
	core.MustNewAPIRoute(`courses/users/enroll`, HandleEnroll),
	core.MustNewAPIRoute(`courses/users/get`, HandleGet),
	core.MustNewAPIRoute(`courses/users/list`, HandleList),
}

func GetRoutes() *[]core.Route {
	return &routes
}
