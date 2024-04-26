package report

// All the API endpoints handled by this package.

import (
    "github.com/edulinq/autograder/api/core"
)

var routes []*core.Route = []*core.Route{
    core.NewAPIRoute(core.NewEndpoint(`report/submissions/fetch`), HandleFetchCourseReport),

};

func GetRoutes() *[]*core.Route {
    return &routes;
}
