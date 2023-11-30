package admin

import (
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
    core.APITestingMain(suite, GetRoutes());
}
