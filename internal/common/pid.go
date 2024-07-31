package common

import (
	"os"
	"strconv"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
)

var pidFilePath = config.PID_PATH.Get()

func CreatePIDFile() error {
	_, err := os.Stat(pidFilePath);

	if (err == nil) {
		log.Fatal("Another instance of the autograder server is already running")

		// data, err := os.ReadFile(pidFilePath);
		// if (err != nil) {
		// 	return fmt.Errorf("Could not read PID file");
		// }

		// if (len(data) > 0) {
		// 	return fmt.Errorf("Another instance of the autograder server is already running")
		// }
	}

	pid := os.Getpid();
	err = os.WriteFile(pidFilePath, []byte(strconv.Itoa(pid)), 0644);
	if (err != nil) {
		log.Error("Failed to write to the PID file.", err);
	}

	return nil;
}

func RemovePIDFile() error {
	err := os.Remove(pidFilePath);
	if (err != nil) {
		log.Error("Failed to remove the PID file.", err);
	}

	return nil;
}