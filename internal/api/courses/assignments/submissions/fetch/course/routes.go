package course

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.MustNewAPIRoute(`courses/assignments/submissions/fetch/course/attempts`, HandleFetchCourseAttempts),
	core.MustNewAPIRoute(`courses/assignments/submissions/fetch/course/scores`, HandleFetchCourseScores),
}

func GetRoutes() *[]core.Route {
	return &routes
}
