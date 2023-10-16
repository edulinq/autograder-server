package submission

// All the API endpoints handled by this package.

import (
    "github.com/eriq-augustine/autograder/api/core"
)

var routes []*core.Route = []*core.Route{
    core.NewAPIRoute(core.NewEndpoint(`submission/peek`), HandlePeek),
    core.NewAPIRoute(core.NewEndpoint(`submission/fetch/scores`), HandleFetchScores),
};

func GetRoutes() *[]*core.Route {
    return &routes;
}
