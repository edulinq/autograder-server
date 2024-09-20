package util

import (
	"testing"
)

func TestVersion(test *testing.T) {
	version := GetAutograderVersion()
	if version.Short == UNKNOWN_VERSION {
		test.Fatalf("Did not get an actual version (check error log).")
	}
}
