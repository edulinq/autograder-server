package common

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

func CreatePIDFile() error {
	var pidFilePath = config.GetPidDir()

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

func CheckAndHandlePIDFile(pidFilePath string) bool {
	if util.IsFile(pidFilePath) {
		data, err := util.ReadFile(pidFilePath)
        if err != nil {
            log.Warn("Failed to read PID file.", err)
            return true
        }

        pid, err := strconv.Atoi(data)
        if err != nil {
            log.Warn("PID value is a string '%s', but could not be converted to an int: %w.", data, err)
            return true
        }

        process, err := os.FindProcess(pid)
        if err != nil {
            log.Warn("Failed to find process.", err)
            return true
        }

        err = process.Signal(syscall.Signal(0))
        if err != nil {
			log.Warn("Removing stale PID file.", err)
			util.RemoveDirent(pidFilePath)
			return true
        } else {
            log.Warn("Another instance of the autograder server is already running.")
            return false
        }
	}

    return true
}