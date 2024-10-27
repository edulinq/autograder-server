package log

import (
	"os"
)

var (
	lastExitCode         = 0
	ShouldExitForTesting = true
)

// To properly test fatal log errors, we need to capture the last exit code.
// Set ShouldExitForTesting to false and compare the last exit code.
// This function ONLY exists to avoid a dependency cycles and should only be called within log.
// EVERYONE should use util.Exit().
func ExitForUtil(code int) {
	Debug("Exiting.", NewAttr("code", code))

	lastExitCode = code

	if ShouldExitForTesting {
		os.Exit(code)
	}
}

func GetLastExitCode() int {
	return lastExitCode
}
