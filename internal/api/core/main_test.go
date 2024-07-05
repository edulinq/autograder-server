package core

import (
	"testing"
)

// This package is infrastructure, and therefore doesn't have any exposed routes.
// However, this slice will be used for runing tests.
// Routes can be added to it during a test and it will be picked up.
var routes []*Route = make([]*Route, 0)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
	APITestingMain(suite, &routes)
}
