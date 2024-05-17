package user

import (
    "testing"

    "github.com/edulinq/autograder/internal/api/core"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
    core.APITestingMain(suite, GetRoutes());
}
