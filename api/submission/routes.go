package submission

// Routes for submission operations.

import (
    "github.com/eriq-augustine/autograder/api/core"
)

var routes []*core.Route = []*core.Route{
    core.NewAPIRoute(core.NewEndpoint(`submission/peek`), HandlePeek),
};

func GetRoutes() *[]*core.Route {
    return &routes;
}
