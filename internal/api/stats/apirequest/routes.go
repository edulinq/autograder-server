package apirequest

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.MustNewAPIRoute(`stats/apirequest/query`, HandleQuery),
}

func GetRoutes() *[]core.Route {
	return &routes
}
