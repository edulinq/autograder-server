package common

import (
	"os"
	"strconv"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)


func CreatePIDFile() error {
	var pidFilePath = config.PID_PATH.Get()

	exists := util.IsFile(pidFilePath)
	if exists {
		log.Fatal("Another instance of the autograder server is already running")
	}

	pid := os.Getpid()
	pidString := strconv.Itoa(pid)

	err := util.WriteFile(pidString, pidFilePath)
	if err != nil {
		log.Error("Failed to write to the PID file.", err)
	}



	return nil
}
