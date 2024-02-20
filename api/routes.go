package api

// All routes handled by the server.

import (
    "github.com/edulinq/autograder/api/admin"
    "github.com/edulinq/autograder/api/core"
    "github.com/edulinq/autograder/api/lms"
    "github.com/edulinq/autograder/api/submission"
    "github.com/edulinq/autograder/api/user"
)

var baseRoutes = []*core.Route{
    core.NewRedirect("GET", ``, `/static/index.html`),
    core.NewRedirect("GET", `/`, `/static/index.html`),
    core.NewRedirect("GET", `/index.html`, `/static/index.html`),

    core.NewRoute("GET", `/static`, handleStatic),
    core.NewRoute("GET", `/static/.*`, handleStatic),
}

func GetRoutes() *[]*core.Route {
    routes := make([]*core.Route, 0);

    routes = append(routes, baseRoutes...);
    routes = append(routes, *(lms.GetRoutes())...);
    routes = append(routes, *(user.GetRoutes())...);
    routes = append(routes, *(submission.GetRoutes())...);
    routes = append(routes, *(admin.GetRoutes())...);

    return &routes;
}
