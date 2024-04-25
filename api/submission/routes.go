package submission

// All the API endpoints handled by this package.

import (
    "github.com/edulinq/autograder/api/core"
)

var routes []*core.Route = []*core.Route{
    core.NewAPIRoute(core.NewEndpoint(`submission/history`), HandleHistory),
    core.NewAPIRoute(core.NewEndpoint(`submission/peek`), HandlePeek),
    core.NewAPIRoute(core.NewEndpoint(`submission/fetch/attempts`), HandleFetchAttempts),
    core.NewAPIRoute(core.NewEndpoint(`submission/fetch/course-report`), HandleFetchCourseReport),
    core.NewAPIRoute(core.NewEndpoint(`submission/fetch/scores`), HandleFetchScores),
    core.NewAPIRoute(core.NewEndpoint(`submission/fetch/submission`), HandleFetchSubmission),
    core.NewAPIRoute(core.NewEndpoint(`submission/fetch/submissions`), HandleFetchSubmissions),
    core.NewAPIRoute(core.NewEndpoint(`submission/submit`), HandleSubmit),
    core.NewAPIRoute(core.NewEndpoint(`submission/remove`), HandleRemoveSubmission),
};

func GetRoutes() *[]*core.Route {
    return &routes;
}
