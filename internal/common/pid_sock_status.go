package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/util"
)

const (
	PID_FILENAME         = "autograder.pid"
	UNIX_SOCKET_FILENAME = "autograder.sock"
	STATUS_FILENAME      = "status.json"
	UNIX_SOCKET_RANDNUM  = 32
	PID_SOCK_LOCK        = "pid_sock_lock"
)

type StatusInfo struct {
	Pid            int    `json:"pid"`
	UnixSocketPath string `json:"unix_socket_path"`
}

func GetStatusPath() string {
	return filepath.Join(config.GetWorkDir(), STATUS_FILENAME)
}

func WriteAndHandlePidStatus() error {
	Lock(PID_SOCK_LOCK)
	defer Unlock(PID_SOCK_LOCK)

	statusPath := GetStatusPath()
	pid := os.Getpid()
	var statusJson StatusInfo

	if !CheckAndHandlePid() {
		return fmt.Errorf("Failed to check and handle the pid.")
	}

	statusJson.Pid = pid

	err := util.ToJSONFile(statusJson, statusPath)
	if err != nil {
		return fmt.Errorf("Failed to write the pid.")
	}

	return nil
}

func GetUnixSocketPath() (string, error) {
	Lock(PID_SOCK_LOCK)
	defer Unlock(PID_SOCK_LOCK)

	statusPath := GetStatusPath()
	var statusJson StatusInfo

	if util.IsFile(statusPath) {
		err := util.JSONFromFile(statusPath, &statusJson)
		if err != nil {
			return "", fmt.Errorf("Failed to read the existing status file.")
		}

		if statusJson.UnixSocketPath != "" {
			return statusJson.UnixSocketPath, nil
		}
	}

	unixFileNumber, err := util.RandHex(UNIX_SOCKET_RANDNUM)
	if err != nil {
		return "", fmt.Errorf("Failed to generate a random number for the unix socket path.")
	}

	unixSocketPath := filepath.Join("/tmp", fmt.Sprintf("autograder-%s.sock", unixFileNumber))
	statusJson.UnixSocketPath = unixSocketPath

	err = util.ToJSONFile(statusJson, statusPath)
	if err != nil {
		return "", fmt.Errorf("Failed to write the unix socket path to the status file.")
	}

	return unixSocketPath, nil
}
