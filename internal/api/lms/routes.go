package lms

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/lms/upload"
	"github.com/edulinq/autograder/internal/api/lms/user"
)

func GetRoutes() *[]core.Route {
	routes := make([]core.Route, 0)

	routes = append(routes, *(upload.GetRoutes())...)
	routes = append(routes, *(user.GetRoutes())...)

	return &routes
}
