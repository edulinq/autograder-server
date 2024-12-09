package fetch

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/assignments/submissions/fetch/course"
	"github.com/edulinq/autograder/internal/api/courses/assignments/submissions/fetch/user"
)

func GetRoutes() *[]core.Route {
	routes := make([]core.Route, 0)

	routes = append(routes, *(course.GetRoutes())...)
	routes = append(routes, *(user.GetRoutes())...)

	return &routes
}
