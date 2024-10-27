package log

import (
	"os"
)

var (
	lastExitCode         = 0
	ShouldExitForTesting = true
)

func Exit(code int) {
	Debug("Exiting.", NewAttr("code", code))

	lastExitCode = code

	if ShouldExitForTesting {
		os.Exit(code)
	}
}

func GetLastExitCode() int {
	return lastExitCode
}
