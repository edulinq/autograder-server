package upload

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.MustNewAPIRoute(`lms/upload/scores`, HandleUploadScores),
}

func GetRoutes() *[]core.Route {
	return &routes
}
