package stats

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/assignments"
	cusers "github.com/edulinq/autograder/internal/api/courses/users"
	"github.com/edulinq/autograder/internal/api/users"
)

// Use the common main for all tests in this package.
// Include routes from other packages to test API request metrics across different endpoints.
func TestMain(suite *testing.M) {
	routes := []core.Route{}

	routes = append(routes, *GetRoutes()...)
	routes = append(routes, *(assignments.GetRoutes())...)
	routes = append(routes, *(cusers.GetRoutes())...)
	routes = append(routes, *(users.GetRoutes())...)

	core.APITestingMain(suite, &routes)
}
