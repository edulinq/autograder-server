package password

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.MustNewAPIRoute(`users/password/change`, HandlePasswordChange),
	core.MustNewAPIRoute(`users/password/reset`, HandlePasswordReset),
}

func GetRoutes() *[]core.Route {
	return &routes
}
