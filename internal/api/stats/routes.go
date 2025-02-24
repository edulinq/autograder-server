package stats

// All routes handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/stats/api"
	"github.com/edulinq/autograder/internal/api/stats/system"
)

func GetRoutes() *[]core.Route {
	routes := make([]core.Route, 0)

	routes = append(routes, *(system.GetRoutes())...)
	routes = append(routes, *(api.GetRoutes())...)

	return &routes
}
