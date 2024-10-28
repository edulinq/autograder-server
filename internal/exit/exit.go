package exit

import (
	"os"
)

// Well-known exit codes (sysexits.h).
const (
	EXIT_CONFIG = 78
)

var (
	lastExitCode         = 0
	shouldExitForTesting = true
)

// Everyone should use Exit() instead of exiting via something like os.Exit().
// This allows for easier testing.
func Exit(code int) {
	lastExitCode = code

	if shouldExitForTesting {
		os.Exit(code)
	}
}

func GetLastExitCode() int {
	return lastExitCode
}

func SetShouldExitForTesting(status bool) {
	shouldExitForTesting = status
}
