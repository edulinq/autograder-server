package common

import (
	"os"
	"syscall"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

func CheckAndHandlePid() bool {
	statusPath := GetStatusPath()

	if util.IsFile(statusPath) {
		var statusJson StatusInfo

		err := util.JSONFromFile(statusPath, &statusJson)
		if err != nil {
			log.Error("Failed to convert file to json.", err)
			return false
		}

		jsonPid := statusJson.Pid
		if jsonPid == 0 {
			return true
		}

		process, _ := os.FindProcess(jsonPid)
		err = process.Signal(syscall.Signal(0))
		if err != nil {
			log.Warn("Removing stale pid.")
			statusJson.Pid = 0
			return true
		} else {
			log.Warn("Another instance of the autograder server is already running.")
			return false
		}
	}

	return true
}
