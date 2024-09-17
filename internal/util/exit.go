package util

import (
	"os"

	"github.com/edulinq/autograder/internal/log"
)

var (
	ShouldExit = true
	LastExitCode = 0
)

func Exit(code int) {
	log.Debug("Exiting with code.", log.NewAttr("code", code))
	LastExitCode = code

	if ShouldExit {
		os.Exit(code)
	}
}

func GetLastExitCode() int {
	return LastExitCode
}