package users

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []*core.Route = []*core.Route{
	core.NewAPIRoute(core.NewEndpoint(`users/auth`), HandleAuth),
	core.NewAPIRoute(core.NewEndpoint(`users/get`), HandleGet),
	core.NewAPIRoute(core.NewEndpoint(`users/list`), HandleList),
	core.NewAPIRoute(core.NewEndpoint(`users/password/change`), HandlePasswordChange),
	core.NewAPIRoute(core.NewEndpoint(`users/password/reset`), HandlePasswordReset),
	core.NewAPIRoute(core.NewEndpoint(`users/tokens/create`), HandleTokensCreate),
	core.NewAPIRoute(core.NewEndpoint(`users/tokens/delete`), HandleTokensDelete),
}

func GetRoutes() *[]*core.Route {
	return &routes
}
