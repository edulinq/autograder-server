package submissions

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []*core.Route = []*core.Route{
	core.NewAPIRoute(core.NewEndpoint(`submissions/history`), HandleHistory),
	core.NewAPIRoute(core.NewEndpoint(`submissions/peek`), HandlePeek),
	core.NewAPIRoute(core.NewEndpoint(`submissions/fetch/attempts`), HandleFetchAttempts),
	core.NewAPIRoute(core.NewEndpoint(`submissions/fetch/scores`), HandleFetchScores),
	core.NewAPIRoute(core.NewEndpoint(`submissions/fetch/submission`), HandleFetchSubmission),
	core.NewAPIRoute(core.NewEndpoint(`submissions/fetch/submissions`), HandleFetchSubmissions),
	core.NewAPIRoute(core.NewEndpoint(`submissions/submit`), HandleSubmit),
	core.NewAPIRoute(core.NewEndpoint(`submissions/remove`), HandleRemoveSubmission),
}

func GetRoutes() *[]*core.Route {
	return &routes
}
