package assignments

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/assignments/submissions"
)

var routes []core.Route = []core.Route{
	core.MustNewAPIRoute(`courses/assignments/get`, HandleGet),
	core.MustNewAPIRoute(`courses/assignments/list`, HandleList),
}

func GetRoutes() *[]core.Route {
	fullRoutes := append(routes, *(submissions.GetRoutes())...)
	return &fullRoutes
}
