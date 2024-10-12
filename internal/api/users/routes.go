package users

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.NewAPIRoute(`users/auth`, HandleAuth, "Authenticate as a user."),
	core.NewAPIRoute(`users/get`, HandleGet, "Get the information for a server user."),
	core.NewAPIRoute(`users/list`, HandleList, "List the users on the server."),
	core.NewAPIRoute(`users/password/change`, HandlePasswordChange, "Change your password to the one provided."),
	core.NewAPIRoute(`users/password/reset`, HandlePasswordReset, "Reset to a random password that will be emailed to you."),
	core.NewAPIRoute(`users/remove`, HandleRemove, "Remove a user from the server."),
	core.NewAPIRoute(`users/tokens/create`, HandleTokensCreate, "Create a new authentication token."),
	core.NewAPIRoute(`users/tokens/delete`, HandleTokensDelete, "Delete an authentication token."),
	core.NewAPIRoute(`users/upsert`, HandleUpsert, "Upsert one or more users to the server (update if exists, insert otherwise)."),
}

func GetRoutes() *[]core.Route {
	return &routes
}
