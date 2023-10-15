package user

// Routes for user operations.

import (
    "github.com/eriq-augustine/autograder/api/core"
)

var routes []*core.Route = []*core.Route{
    core.NewAPIRoute(core.NewEndpoint(`user/get`), HandleUserGet),
    core.NewAPIRoute(core.NewEndpoint(`user/list`), HandleUserList),
};

func GetRoutes() *[]*core.Route {
    return &routes;
}
