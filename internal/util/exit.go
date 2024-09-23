package util

import (
	"os"

	"github.com/edulinq/autograder/internal/log"
)

var (
	lastExitCode         = 0
	ShouldExitForTesting = true
)

func Exit(code int) {
	log.Debug("Exiting.", log.NewAttr("code", code))

	lastExitCode = code

	if ShouldExitForTesting {
		os.Exit(code)
	}
}

func GetLastExitCode() int {
	return lastExitCode
}
