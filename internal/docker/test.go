package docker

import (
	"testing"

	"github.com/edulinq/autograder/internal/config"
)

func EnsureOrSkipForTest(test *testing.T) {
	if config.DOCKER_DISABLE.Get() {
		test.Skip("Docker is disabled, skipping test.")
	}

	if !CanAccessDocker() {
		test.Fatal("Could not access docker.")
	}
}
