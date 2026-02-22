package tokens

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.MustNewAPIRoute(`users/tokens/create`, HandleCreate),
	core.MustNewAPIRoute(`users/tokens/delete`, HandleDelete),
	core.MustNewAPIRoute(`users/tokens/list`, HandleList),
}

func GetRoutes() *[]core.Route {
	return &routes
}
