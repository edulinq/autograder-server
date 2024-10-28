package upsert

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.MustNewAPIRoute(`courses/upsert/filespec`, HandleFileSpec),
	core.MustNewAPIRoute(`courses/upsert/zip`, HandleZipFile),
}

func GetRoutes() *[]core.Route {
	return &routes
}
