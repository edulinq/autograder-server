package lms

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []*core.Route = []*core.Route{
	core.NewAPIRoute(core.NewEndpoint(`lms/user/get`), HandleUserGet),
	core.NewAPIRoute(core.NewEndpoint(`lms/sync`), HandleSync),
	core.NewAPIRoute(core.NewEndpoint(`lms/upload/scores`), HandleUploadScores),
}

func GetRoutes() *[]*core.Route {
	return &routes
}
