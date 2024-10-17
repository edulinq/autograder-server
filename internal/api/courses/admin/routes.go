package admin

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.NewAPIRoute(`courses/admin/update`, HandleUpdate),
}

func GetRoutes() *[]core.Route {
	return &routes
}
