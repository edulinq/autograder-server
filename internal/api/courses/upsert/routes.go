package upsert

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []*core.Route = []*core.Route{
	core.NewAPIRoute(`courses/upsert/filespec`, HandleFileSpec),
	core.NewAPIRoute(`courses/upsert/zip`, HandleZipFile),
}

func GetRoutes() *[]*core.Route {
	return &routes
}
