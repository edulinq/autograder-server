package lms

// All routes handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/lms/scores"
)

func GetRoutes() *[]core.Route {
	return scores.GetRoutes()
}
