package common

import (
	"fmt"
	"os"
	"strconv"
)

const (
	pidFilePath = "/tmp/autograder.pid"
)

func CreatePIDFile() error {
	_, err := os.Stat(pidFilePath);

	if (err == nil) {
		return fmt.Errorf("Another instance of the autograder server is already running")

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
		return fmt.Errorf("Error writing PID file. ", err);
	}

	return nil;
}

func RemovePIDFile() error {
	err := os.Remove(pidFilePath);
	if (err != nil) {
		return fmt.Errorf("Could not remove PID file");
	}

	return nil;
}