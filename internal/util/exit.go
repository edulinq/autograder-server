package util

import (
	"os"

	"github.com/edulinq/autograder/internal/log"
)

var (
	ShouldExit   = true
	LastExitCode = 0
)

func Exit(code int) {
	SetExitCode(code)

	if ShouldExit {
		log.Debug("Exiting with code.", log.NewAttr("code", code))
		os.Exit(code)
	}
}

func GetLastExitCode() int {
	return LastExitCode
}

func SetExitCode(code int) {
	LastExitCode = code
}
