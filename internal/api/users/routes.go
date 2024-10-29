package users

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.MustNewAPIRoute(`users/auth`, HandleAuth),
	core.MustNewAPIRoute(`users/get`, HandleGet),
	core.MustNewAPIRoute(`users/list`, HandleList),
	core.MustNewAPIRoute(`users/password/change`, HandlePasswordChange),
	core.MustNewAPIRoute(`users/password/reset`, HandlePasswordReset),
	core.MustNewAPIRoute(`users/remove`, HandleRemove),
	core.MustNewAPIRoute(`users/tokens/create`, HandleTokensCreate),
	core.MustNewAPIRoute(`users/tokens/delete`, HandleTokensDelete),
	core.MustNewAPIRoute(`users/upsert`, HandleUpsert),
}

func GetRoutes() *[]core.Route {
	return &routes
}
