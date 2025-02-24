package api

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.MustNewAPIRoute(`stats/api/query`, HandleQuery),
}

func GetRoutes() *[]core.Route {
	return &routes
}
