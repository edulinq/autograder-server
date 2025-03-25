package util

import (
	"fmt"
	"strings"
	"testing"
)

func TestVersion(test *testing.T) {
	version, err := GetFullCachedVersion()
	if err != nil {
		test.Fatalf("Failed to get autograder version: %s", err)
	}

	if strings.HasPrefix(version.Base, fmt.Sprintf("%d", UNKNOWN_COMPONENT)) {
		test.Fatalf("Did not get an actual version (check error log).")
	}
}

func TestMustGetAPIVersion(test *testing.T) {
	apiVersion := MustGetAPIVersion()

	if apiVersion == UNKNOWN_COMPONENT {
		test.Fatalf("Did not get an actual version (check error log).")
	}
}
