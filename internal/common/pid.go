package common

import (
	"os"
	"syscall"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

func CheckAndHandleStalePid() bool {
	statusPath := GetStatusPath()

	if !util.IsFile(statusPath) {
		return true
	}

	var statusJson StatusInfo

	err := util.JSONFromFile(statusPath, &statusJson)
	if err != nil {
		log.Error("Failed to read the existing status file.", err)
		return false
	}

	if statusJson.Pid == 0 {
		return true
	}

	process, _ := os.FindProcess(statusJson.Pid)
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		log.Warn("Removing stale status file.")
		util.RemoveDirent(GetStatusPath())
		return true
	} else {
		return false
	}
}
