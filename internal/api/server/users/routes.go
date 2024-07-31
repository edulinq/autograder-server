package users

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []*core.Route = []*core.Route{
	core.NewAPIRoute(core.NewEndpoint(`server/users/auth`), HandleAuth),
	core.NewAPIRoute(core.NewEndpoint(`server/users/get`), HandleGet),
	core.NewAPIRoute(core.NewEndpoint(`server/users/list`), HandleList),
	core.NewAPIRoute(core.NewEndpoint(`server/users/tokens/create`), HandleTokensCreate),
	core.NewAPIRoute(core.NewEndpoint(`server/users/tokens/delete`), HandleTokensDelete),
}

func GetRoutes() *[]*core.Route {
	return &routes
}
