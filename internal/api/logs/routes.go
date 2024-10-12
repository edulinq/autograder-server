package logs

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.NewAPIRoute(`logs/query`, HandleQuery, "Query log entries from the autograder server."),
}

func GetRoutes() *[]core.Route {
	return &routes
}
