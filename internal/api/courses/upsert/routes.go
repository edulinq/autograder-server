package upsert

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.NewAPIRoute(`courses/upsert/filespec`, HandleFileSpec, "Upsert a course using a filespec."),
	core.NewAPIRoute(`courses/upsert/zip`, HandleZipFile, "Upsert a course using a zip file."),
}

func GetRoutes() *[]core.Route {
	return &routes
}
