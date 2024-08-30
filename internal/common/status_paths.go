package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/util"
)

const (
	PID_FILENAME         = "grade.pid"
	UNIX_SOCKET_FILENAME = "a.sock"
	STATUS_FILENAME      = "status.json"
	UNIX_SOCKET_RANDNUM  = 32
)

type StatusInfo struct {
	Pid            int    `json:"pid"`
	UnixSocketPath string `json:"unix_socket_path"`
}

func GetStatusPath() string {
	return filepath.Join(config.GetWorkDir(), STATUS_FILENAME)
}

func WriteAndHandlePidStatus() error {
	statusPath := GetStatusPath()
	pid := os.Getpid()

	if !CheckAndHandlePid(pid) {
		return fmt.Errorf("Failed to check and handle the pid.")
	}

	var status StatusInfo
	status.Pid = pid

	err := util.ToJSONFile(status, statusPath)
	if err != nil {
		return fmt.Errorf("Failed to write status to json: %w", err)
	}

	return nil
}

func WriteAndReturnUnixSocketPath() (string, error) {
	statusPath := GetStatusPath()
	var status StatusInfo

	if util.IsFile(statusPath) {
		err := util.JSONFromFile(statusPath, &status)
		if err != nil {
			return "", fmt.Errorf("Failed to read existing status file")
		}

		if status.UnixSocketPath != "" {
			return status.UnixSocketPath, nil
		}
	}

	unixFileNumber, err := util.RandHex(UNIX_SOCKET_RANDNUM)
	if err != nil {
		return "", fmt.Errorf("Failed to generate random number: %w", err)
	}

	socketPath := filepath.Join("/tmp", fmt.Sprintf("autograder-%s.sock", unixFileNumber))
	status.UnixSocketPath = socketPath

	err = util.ToJSONFile(status, statusPath)
	if err != nil {
		return "", fmt.Errorf("Failed to write status to json: %w", err)
	}

	return socketPath, nil
}
