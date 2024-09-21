package util

import (
	"testing"
)

func TestVersion(test *testing.T) {
	version, err := GetAutograderVersion()
	if err != nil {
		test.Fatalf("Failed to get autograder version: %s", err)
	}

	if version.Short == UNKNOWN_VERSION {
		test.Fatalf("Did not get an actual version (check error log).")
	}

	if version.Api == UNKNOWN_API {
		test.Fatalf("Did not get an actual API version (check error log).")
	}
}
