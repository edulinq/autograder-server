package lms

import (
    "testing"

    "github.com/edulinq/autograder/api/core"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
    core.APITestingMain(suite, GetRoutes());
}
