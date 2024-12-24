package users

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/users/password"
	"github.com/edulinq/autograder/internal/api/users/tokens"
)

var baseRoutes []core.Route = []core.Route{
	core.MustNewAPIRoute(`users/auth`, HandleAuth),
	core.MustNewAPIRoute(`users/get`, HandleGet),
	core.MustNewAPIRoute(`users/list`, HandleList),
	core.MustNewAPIRoute(`users/remove`, HandleRemove),
	core.MustNewAPIRoute(`users/upsert`, HandleUpsert),
}

func GetRoutes() *[]core.Route {
	routes := make([]core.Route, 0)

	routes = append(routes, baseRoutes...)
	routes = append(routes, *(password.GetRoutes())...)
	routes = append(routes, *(tokens.GetRoutes())...)

	return &routes
}
