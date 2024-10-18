package submissions

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.NewAPIRoute(`courses/assignments/submissions/fetch/course/attempts`, HandleFetchCourseAttempts),
	core.NewAPIRoute(`courses/assignments/submissions/fetch/course/scores`, HandleFetchCourseScores),

	core.NewAPIRoute(`courses/assignments/submissions/fetch/user/attempt`, HandleFetchUserAttempt),
	core.NewAPIRoute(`courses/assignments/submissions/fetch/user/attempts`, HandleFetchUserAttempts),
	core.NewAPIRoute(`courses/assignments/submissions/fetch/user/history`, HandleFetchUserHistory),
	core.NewAPIRoute(`courses/assignments/submissions/fetch/user/peek`, HandleFetchUserPeek),

	core.NewAPIRoute(`courses/assignments/submissions/remove`, HandleRemove),
	core.NewAPIRoute(`courses/assignments/submissions/submit`, HandleSubmit),
}

func GetRoutes() *[]core.Route {
	return &routes
}
