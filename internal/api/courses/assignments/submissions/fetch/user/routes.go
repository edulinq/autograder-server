package user

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.MustNewAPIRoute(`courses/assignments/submissions/fetch/user/attempt`, HandleFetchUserAttempt),
	core.MustNewAPIRoute(`courses/assignments/submissions/fetch/user/attempts`, HandleFetchUserAttempts),
	core.MustNewAPIRoute(`courses/assignments/submissions/fetch/user/history`, HandleFetchUserHistory),
	core.MustNewAPIRoute(`courses/assignments/submissions/fetch/user/peek`, HandleFetchUserPeek),
}

func GetRoutes() *[]core.Route {
	return &routes
}
