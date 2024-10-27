package util

import (
	"github.com/edulinq/autograder/internal/log"
)

// Everyone should use Exit() for coherent code.
func Exit(code int) {
	log.ExitForUtil(code)
}

func GetLastExitCode() int {
	return log.GetLastExitCode()
}
