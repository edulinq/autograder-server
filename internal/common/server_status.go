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

type ServerInitiator string

const (
	SERVER_STATUS_LOCK             = "internal.common.SERVER_STATUS_LOCK"
	STATUS_FILENAME                = "status.json"
	UNIX_SOCKET_RANDNUM_SIZE_BYTES = 32
)

const (
	PRIMARY_SERVER  ServerInitiator = "primary-server"
	CMD_SERVER      ServerInitiator = "cmd-server"
	CMD_TEST_SERVER ServerInitiator = "cmd-test-server"
)

type StatusInfo struct {
	Pid             int             `json:"pid"`
	UnixSocketPath  string          `json:"unix-socket-path"`
	ServerInitiator ServerInitiator `json:"server-initiator"`
}

func GetStatusPath() string {
	return filepath.Join(config.GetWorkDir(), STATUS_FILENAME)
}

func GetUnixSocketPath() (string, error) {
	ReadLock(SERVER_STATUS_LOCK)
	defer ReadUnlock(SERVER_STATUS_LOCK)

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

func WriteAndHandleStatusFile(initiator ServerInitiator) error {
	Lock(SERVER_STATUS_LOCK)
	Unlock(SERVER_STATUS_LOCK)

	statusPath := GetStatusPath()
	pid := os.Getpid()

	statusInfo, err := CheckAndHandleServerStatusFile()
	if err != nil {
		return err
	}

	if statusInfo != nil {
		return fmt.Errorf("Failed to create the status file '%s' since another server is running.", statusPath)
	}

	unixFileNumber, err := util.RandHex(UNIX_SOCKET_RANDNUM_SIZE_BYTES)
	if err != nil {
		return fmt.Errorf("Failed to generate a random number for the unix socket path: '%w'.", err)
	}

	statusJson := StatusInfo{
		Pid:             pid,
		UnixSocketPath:  filepath.Join("/", "tmp", fmt.Sprintf("autograder-%s.sock", unixFileNumber)),
		ServerInitiator: initiator,
	}

	err = util.ToJSONFile(statusJson, statusPath)
	if err != nil {
		return fmt.Errorf("Failed to write to the status file '%s': '%w'.", statusPath, err)
	}

	return nil
}

// Returns (nil, nil) if the status file doesn't exist,
// (&statusJson, nil) if the status file exists and another instance of the server is running,
// or (nil, err) if there are issues reading or removing the status file.
func CheckAndHandleServerStatusFile() (*StatusInfo, error) {
	statusPath := GetStatusPath()
	if !util.IsFile(statusPath) {
		return nil, nil
	}

	var statusJson StatusInfo
	err := util.JSONFromFile(statusPath, &statusJson)
	if err != nil {
		return nil, fmt.Errorf("Failed to read the status file '%s': '%w'.", statusPath, err)
	}

	if isAlive(statusJson.Pid) {
		return &statusJson, nil
	} else {
		log.Warn("Removing stale status file.", log.NewAttr("path", statusPath))

		err := util.RemoveDirent(statusPath)
		if err != nil {
			return nil, fmt.Errorf("Failed to remove the status file '%s': '%w'.", statusPath, err)
		}
	}

	return nil, nil
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
