package lms

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.NewAPIRoute(`lms/user/get`, HandleUserGet, "Get information for an LMS user."),
	core.NewAPIRoute(`lms/upload/scores`, HandleUploadScores,
		"Upload scores from a tab-separated file to the course's LMS. The file should not have headers, and should have two columns: email and score."),
}

func GetRoutes() *[]core.Route {
	return &routes
}
