package metadata

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.MustNewAPIRoute(`metadata/describe`, HandleDescribe),
	core.MustNewAPIRoute(`metadata/heartbeat`, HandleHeartbeat),
}

func GetRoutes() *[]core.Route {
	return &routes
}
