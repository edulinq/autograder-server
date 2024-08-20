package common

import (
	"fmt"
	"os"
	"strconv"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/util"
)

func CreatePIDFile() error {
	var pidFilePath = config.PID_PATH.Get()

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
