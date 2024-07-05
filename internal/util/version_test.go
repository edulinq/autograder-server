package util

import (
	"strings"
	"testing"
)

func TestVersionShort(test *testing.T) {
	version := GetAutograderVersion()
	if version == UNKNOWN_VERSION {
		test.Fatalf("Did not get an actual version (check error log).")
	}
}

func TestVersionFull(test *testing.T) {
	version := GetAutograderFullVersion()
	if strings.HasPrefix(version, UNKNOWN_VERSION) {
		test.Fatalf("Did not get an actual version (check error log).")
	}
}
