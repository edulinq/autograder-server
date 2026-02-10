package assignments

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/assignments/images"
	"github.com/edulinq/autograder/internal/api/courses/assignments/submissions"
)

var baseRoutes []core.Route = []core.Route{
	core.MustNewAPIRoute(`courses/assignments/get`, HandleGet),
	core.MustNewAPIRoute(`courses/assignments/list`, HandleList),
	core.MustNewAPIRoute(`courses/assignments/report`, HandleCourseReport),
}

func GetRoutes() *[]core.Route {
	routes := make([]core.Route, 0)

	routes = append(routes, baseRoutes...)
	routes = append(routes, *(images.GetRoutes())...)
	routes = append(routes, *(submissions.GetRoutes())...)

	return &routes
}
