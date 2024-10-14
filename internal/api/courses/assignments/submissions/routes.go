package submissions

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []*core.Route = []*core.Route{
	core.NewAPIRoute(core.NewEndpoint(`courses/assignments/submissions/fetch/course/attempts`), HandleFetchCourseAttempts),
	core.NewAPIRoute(core.NewEndpoint(`courses/assignments/submissions/fetch/course/scores`), HandleFetchCourseScores),

	core.NewAPIRoute(core.NewEndpoint(`courses/assignments/submissions/fetch/user/attempt`), HandleFetchUserAttempt),
	core.NewAPIRoute(core.NewEndpoint(`courses/assignments/submissions/fetch/user/attempts`), HandleFetchUserAttempts),
	core.NewAPIRoute(core.NewEndpoint(`courses/assignments/submissions/fetch/user/history`), HandleFetchUserHistory),
	core.NewAPIRoute(core.NewEndpoint(`courses/assignments/submissions/fetch/user/peek`), HandleFetchUserPeek),

	core.NewAPIRoute(core.NewEndpoint(`courses/assignments/submissions/submit`), HandleSubmit),
	core.NewAPIRoute(core.NewEndpoint(`courses/assignments/submissions/remove`), HandleRemove),
}

func GetRoutes() *[]*core.Route {
	return &routes
}
