package users

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.NewAPIRoute(`users/auth`, HandleAuth),
	core.NewAPIRoute(`users/get`, HandleGet),
	core.NewAPIRoute(`users/list`, HandleList),
	core.NewAPIRoute(`users/password/change`, HandlePasswordChange),
	core.NewAPIRoute(`users/password/reset`, HandlePasswordReset),
	core.NewAPIRoute(`users/remove`, HandleRemove),
	core.NewAPIRoute(`users/tokens/create`, HandleTokensCreate),
	core.NewAPIRoute(`users/tokens/delete`, HandleTokensDelete),
	core.NewAPIRoute(`users/upsert`, HandleUpsert),
}

func GetRoutes() *[]core.Route {
	return &routes
}
