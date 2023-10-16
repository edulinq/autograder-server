package lms

// All the API endpoints handled by this package.

import (
    "github.com/eriq-augustine/autograder/api/core"
)

var routes []*core.Route = []*core.Route{
    core.NewAPIRoute(core.NewEndpoint(`lms/user/get`), HandleUserGet),
};

func GetRoutes() *[]*core.Route {
    return &routes;
}
