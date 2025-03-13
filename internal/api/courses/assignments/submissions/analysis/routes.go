package analysis

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.MustNewAPIRoute(`courses/assignments/submissions/analysis/individual`, HandleIndividual),
	core.MustNewAPIRoute(`courses/assignments/submissions/analysis/pairwise`, HandlePairwise),
}

func GetRoutes() *[]core.Route {
	return &routes
}
