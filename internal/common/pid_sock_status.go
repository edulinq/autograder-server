package common

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

const (
	PID_SOCK_LOCK                  = "PID_SOCK_LOCK"
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

func WriteAndHandleStatusFile() (err error) {
	Lock(PID_SOCK_LOCK)
	defer func() {
		err = errors.Join(err, Unlock(PID_SOCK_LOCK))
	}()

	statusPath := GetStatusPath()
	pid := os.Getpid()
	var statusJson StatusInfo

	ok, err := checkAndHandleStalePid()
	if !ok {
		if err != nil {
			return err
		}

		return fmt.Errorf("Failed to create the status file.")
	}

	statusJson.Pid = pid

	unixFileNumber, err := util.RandHex(UNIX_SOCKET_RANDNUM_SIZE_BYTES)
	if err != nil {
		return fmt.Errorf("Failed to generate a random number for the unix socket path: '%w'.", err)
	}
	statusJson.UnixSocketPath = filepath.Join("/tmp", fmt.Sprintf("autograder-%s.sock", unixFileNumber))

	err = util.ToJSONFile(statusJson, statusPath)
	if err != nil {
		return fmt.Errorf("Failed to write the pid to the status file: '%w'.", err)
	}

	return err
}

func checkAndHandleStalePid() (bool, error) {
	statusPath := GetStatusPath()

	if !util.IsFile(statusPath) {
		return true, nil
	}

	var statusJson StatusInfo

	err := util.JSONFromFile(statusPath, &statusJson)
	if err != nil {
		return false, fmt.Errorf("Failed to read the existing status file: '%w'.", err)
	}

	process, _ := os.FindProcess(statusJson.Pid)
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		log.Warn("Removing stale status file.")

		err := util.RemoveDirent(GetStatusPath())
		if err != nil {
			return false, fmt.Errorf("Failed to remove the status file: '%w'.", err)
		}

		return true, nil
	} else {
		return false, nil
	}
}

func GetUnixSocketPath() (path string, err error) {
	ReadLock(PID_SOCK_LOCK)
	defer func() {
		err = errors.Join(err, ReadUnlock(PID_SOCK_LOCK))
	}()

	statusPath := GetStatusPath()
	if !util.IsFile(statusPath) {
		return "", fmt.Errorf("The status path doesn't exist.")
	}

	var statusJson StatusInfo

	err = util.JSONFromFile(statusPath, &statusJson)
	if err != nil {
		return "", fmt.Errorf("Failed to read the existing status file: '%w'.", err)
	}

	if statusJson.UnixSocketPath == "" {
		return "", fmt.Errorf("The unix socket path is empty.")
	}

	return statusJson.UnixSocketPath, nil
}
