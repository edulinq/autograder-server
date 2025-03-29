package submissions

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/assignments/submissions/analysis"
	"github.com/edulinq/autograder/internal/api/courses/assignments/submissions/fetch"
	"github.com/edulinq/autograder/internal/api/courses/assignments/submissions/proxy"
)

var baseRoutes []core.Route = []core.Route{
	core.MustNewAPIRoute(`courses/assignments/submissions/remove`, HandleRemove),
	core.MustNewAPIRoute(`courses/assignments/submissions/submit`, HandleSubmit),
}

func GetRoutes() *[]core.Route {
	routes := make([]core.Route, 0)

	routes = append(routes, baseRoutes...)
	routes = append(routes, *(analysis.GetRoutes())...)
	routes = append(routes, *(fetch.GetRoutes())...)
	routes = append(routes, *(proxy.GetRoutes())...)

	return &routes
}
