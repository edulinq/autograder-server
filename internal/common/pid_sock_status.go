package common

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

const (
	PID_SOCK_LOCK                  = "internal.common.PID_SOCK_LOCK"
	STATUS_FILENAME                = "status.json"
	UNIX_SOCKET_RANDNUM_SIZE_BYTES = 32
)

type StatusInfo struct {
	Pid            int    `json:"pid"`
	UnixSocketPath string `json:"unix_socket_path"`
}

func GetStatusPath() string {
	return filepath.Join(config.GetWorkDir(), STATUS_FILENAME)
}

func GetUnixSocketPath() (string, error) {
	ReadLock(PID_SOCK_LOCK)
	defer ReadUnlock(PID_SOCK_LOCK)

	statusPath := GetStatusPath()
	if !util.IsFile(statusPath) {
		return "", fmt.Errorf("Status file '%s' does not exist.", statusPath)
	}

	var statusJson StatusInfo
	err := util.JSONFromFile(statusPath, &statusJson)
	if err != nil {
		return "", fmt.Errorf("Failed to read the existing status file '%s': '%w'.", statusPath, err)
	}

	if statusJson.UnixSocketPath == "" {
		return "", fmt.Errorf("The unix socket path is empty.")
	}

	return statusJson.UnixSocketPath, nil
}

func WriteAndHandleStatusFile() error {
	Lock(PID_SOCK_LOCK)
	Unlock(PID_SOCK_LOCK)

	statusPath := GetStatusPath()
	pid := os.Getpid()
	var statusJson StatusInfo

	ok, err := checkAndHandleStalePid()
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("Failed to create the status file '%s'.", statusPath)
	}

	statusJson.Pid = pid

	unixFileNumber, err := util.RandHex(UNIX_SOCKET_RANDNUM_SIZE_BYTES)
	if err != nil {
		return fmt.Errorf("Failed to generate a random number for the unix socket path: '%w'.", err)
	}

	statusJson.UnixSocketPath = filepath.Join("/", "tmp", fmt.Sprintf("autograder-%s.sock", unixFileNumber))

	err = util.ToJSONFile(statusJson, statusPath)
	if err != nil {
		return fmt.Errorf("Failed to write to the status file '%s': '%w'.", statusPath, err)
	}

	return nil
}

// Returns (true, nil) if it's safe to create the status file,
// (false, nil) if another instance of the server is running,
// or (false, err) if there are issues reading or removing the status file.
func checkAndHandleStalePid() (bool, error) {
	statusPath := GetStatusPath()

	if !util.IsFile(statusPath) {
		return true, nil
	}

	var statusJson StatusInfo
	err := util.JSONFromFile(statusPath, &statusJson)
	if err != nil {
		return false, fmt.Errorf("Failed to read the status file '%s': '%w'.", statusPath, err)
	}

	if isAlive(statusJson.Pid) {
		return false, nil
	} else {
		log.Warn("Removing stale status file '%s'.", statusPath)

		err := util.RemoveDirent(statusPath)
		if err != nil {
			return false, fmt.Errorf("Failed to remove the status file '%s': '%w'.", statusPath, err)
		}
	}

	return true, nil
}

// Check if the pid is currently being used.
// Returns false if the pid is inactive and true if the pid is active.
func isAlive(pid int) bool {
	process, _ := os.FindProcess(pid)
	err := process.Signal(syscall.Signal(0))
	if err != nil {
		return false
	}

	return true
}
