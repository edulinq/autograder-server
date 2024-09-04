package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/util"
)

const (
	UNIX_SOCKET_FILENAME           = "autograder.sock"
	STATUS_FILENAME                = "status.json"
	UNIX_SOCKET_RANDNUM_SIZE_BYTES = 32
	PID_SOCK_LOCK                  = "PID_SOCK_LOCK"
)

type StatusInfo struct {
	Pid            int    `json:"pid"`
	UnixSocketPath string `json:"unix_socket_path"`
}

func GetStatusPath() string {
	return filepath.Join(config.GetWorkDir(), STATUS_FILENAME)
}

func WriteAndHandleStatusFile() error {
	Lock(PID_SOCK_LOCK)
	defer Unlock(PID_SOCK_LOCK)

	statusPath := GetStatusPath()
	pid := os.Getpid()
	var statusJson StatusInfo

	if !checkAndHandleStalePid() {
		return fmt.Errorf("Failed to check and handle the pid.")
	}

	statusJson.Pid = pid

	unixFileNumber, err := util.RandHex(UNIX_SOCKET_RANDNUM_SIZE_BYTES)
	if err != nil {
		return fmt.Errorf("Failed to generate a random number for the unix socket path: '%w'.", err)
	}
	statusJson.UnixSocketPath = fmt.Sprintf("/tmp/autograder-%s.sock", unixFileNumber)

	err = util.ToJSONFile(statusJson, statusPath)
	if err != nil {
		return fmt.Errorf("Failed to write the pid to the status file: '%w'.", err)
	}

	return nil
}

func GetUnixSocketPath() (string, error) {
	Lock(PID_SOCK_LOCK)
	defer Unlock(PID_SOCK_LOCK)

	statusPath := GetStatusPath()
	if !util.IsFile(statusPath) {
		return "", fmt.Errorf("The status path doesn't exist.")
	}

	var statusJson StatusInfo

	err := util.JSONFromFile(statusPath, &statusJson)
	if err != nil {
		return "", fmt.Errorf("Failed to read the existing status file: '%w'.", err)
	}

	if statusJson.UnixSocketPath != "" {
		return statusJson.UnixSocketPath, nil
	} else {
		return "", fmt.Errorf("The unix socket path is empty.")
	}
}
