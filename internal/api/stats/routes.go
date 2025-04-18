package stats

// All routes handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

func GetRoutes() *[]core.Route {
	routes := make([]core.Route, 0)

	routes = append(routes, core.MustNewAPIRoute(`stats/query`, HandleQuery))

	return &routes
}
