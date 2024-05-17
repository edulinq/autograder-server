package user

// All the API endpoints handled by this package.

import (
    "github.com/edulinq/autograder/internal/api/core"
)

var routes []*core.Route = []*core.Route{
    core.NewAPIRoute(core.NewEndpoint(`user/add`), HandleAdd),
    core.NewAPIRoute(core.NewEndpoint(`user/auth`), HandleAuth),
    core.NewAPIRoute(core.NewEndpoint(`user/change/pass`), HandleChangePassword),
    core.NewAPIRoute(core.NewEndpoint(`user/get`), HandleUserGet),
    core.NewAPIRoute(core.NewEndpoint(`user/list`), HandleList),
    core.NewAPIRoute(core.NewEndpoint(`user/remove`), HandleRemove),
};

func GetRoutes() *[]*core.Route {
    return &routes;
}
