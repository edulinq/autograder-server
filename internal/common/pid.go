package common

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

func CreatePIDFile() error {
	var pidFilePath = GetStatusPath()

	exists := util.IsFile(pidFilePath)
	if exists {
		return fmt.Errorf("Another instance of the autograder server is already running.")
	}

	pid := os.Getpid()
	pidString := strconv.Itoa(pid)

	err := util.WriteFile(pidString, pidFilePath)
	if err != nil {
		return fmt.Errorf("Failed to write to the PID file.")
	}

	return nil
}

func CheckAndHandlePid(pid int) bool {
	statusPath := GetStatusPath()

	if util.IsFile(statusPath) {
		var statusJson StatusInfo
		err := util.JSONFromFile(statusPath, &statusJson)
		if err != nil {
			log.Warn("Failed to convert file to json.")
			return false
		}

		pid := statusJson.Pid

		process, err := os.FindProcess(pid)
		if err != nil {
			log.Warn("Failed to find process.", err)
			return true
		}

		err = process.Signal(syscall.Signal(0))
		if err != nil {
			log.Warn("Removing stale PID file.", err)
			util.RemoveDirent(statusPath)
			return true
		} else {
			log.Warn("Another instance of the autograder server is already running.")
			return false
		}
	}

	return true
}
