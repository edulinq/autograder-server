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

var (
	PrimaryServer = "primary-server"
	CmdServer     = "cmd-server"
	CmdTestServer = "cmd-test-server"
)

type StatusInfo struct {
	Pid            int    `json:"pid"`
	UnixSocketPath string `json:"unix_socket_path"`
	ServerCreator  string `json:"server_creator"`
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

func WriteAndHandleStatusFile(creator string) (bool, error) {
	Lock(PID_SOCK_LOCK)
	Unlock(PID_SOCK_LOCK)

	statusPath := GetStatusPath()
	pid := os.Getpid()
	var statusJson StatusInfo

	ok, err := CheckServerCreator(creator)
	if err != nil {
		return false, err
	}

	if !ok {
		return false, nil
	}

	ok, err = checkAndHandleStalePid()
	if err != nil {
		return false, err
	}

	if !ok {
		return false, fmt.Errorf("Failed to create the status file '%s'.", statusPath)
	}

	statusJson.Pid = pid

	unixFileNumber, err := util.RandHex(UNIX_SOCKET_RANDNUM_SIZE_BYTES)
	if err != nil {
		return false, fmt.Errorf("Failed to generate a random number for the unix socket path: '%w'.", err)
	}

	statusJson.UnixSocketPath = filepath.Join("/", "tmp", fmt.Sprintf("autograder-%s.sock", unixFileNumber))

	statusJson.ServerCreator = creator

	err = util.ToJSONFile(statusJson, statusPath)
	if err != nil {
		return false, fmt.Errorf("Failed to write to the status file '%s': '%w'.", statusPath, err)
	}

	return true, nil
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
		log.Warn("Removing stale status file.", log.NewAttr("path", statusPath))

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

// Returns (true, nil) if it's safe to start a new server,
// (false, nil) if the operation should not start their own server,
// or (false, error) if there are issues with the status file or if multiple cmds try to start servers.
func CheckServerCreator(creator string) (bool, error) {
	// Check if any server is actively running.
	ok, err := checkAndHandleStalePid()
	if err != nil {
		return false, err
	}

	if ok {
		return true, nil
	}

	statusPath := GetStatusPath()
	if !util.IsFile(statusPath) {
		return false, fmt.Errorf("Server is running but status file not found at '%s'.", statusPath)
	}

	var statusJson StatusInfo
	if err := util.JSONFromFile(statusPath, &statusJson); err != nil {
		return false, fmt.Errorf("failed to read status file '%s': '%w'.", statusPath, err)
	}

	if statusJson.ServerCreator == CmdTestServer {
		return false, nil
	}

	// If a cmd is trying to start the server while the primary server is running,
	// have the cmd use the primary server.
	if creator == CmdServer && statusJson.ServerCreator == PrimaryServer {
		log.Info("Connecting to the primary server.")
		return false, nil
	}

	// If a cmd is trying to start the server while a cmd server is running,
	// don't allow the cmd to start a server.
	if creator == CmdServer && statusJson.ServerCreator == CmdServer {
		return false, fmt.Errorf("A CMD has already started a server.")
	}

	return true, nil
}

// Returns (true, nil) if there is an active cmd server running,
// (false, nil) if there isn't a server running with the inputted creator,
// (false, err) if there is an error checking or if another server is running.
// func IsServerCreator(creator string) (bool, error) {
// 	notRunning, err := checkAndHandleStalePid()
// 	if err != nil {
// 		return false, err
// 	}

// 	if !notRunning {
// 		return false, fmt.Errorf("Another server is running.")
// 	}

// 	statusPath := GetStatusPath()

// 	if !util.IsFile(statusPath) {
// 		return false, nil
// 	}

// 	var statusJson StatusInfo
// 	err = util.JSONFromFile(statusPath, &statusJson)
// 	if err != nil {
// 		return false, fmt.Errorf("Failed to read the status file '%s': '%w'.", statusPath, err)
// 	}

// 	if statusJson.ServerCreator == creator {
// 		return true, nil
// 	}

// 	return false, nil
// }

func IsPrimaryServerRunning() (bool, error) {
	// First check if any server is running by checking for stale PIDs
	notRunning, err := checkAndHandleStalePid()
	if err != nil {
		return false, fmt.Errorf("error checking server status: %w", err)
	}

	if notRunning {
		// No server is running
		return false, nil
	}

	// A server is running, now check if it's the primary server
	statusPath := GetStatusPath()
	if !util.IsFile(statusPath) {
		return false, fmt.Errorf("server is running but status file not found at '%s'", statusPath)
	}

	var statusJson StatusInfo
	err = util.JSONFromFile(statusPath, &statusJson)
	if err != nil {
		return false, fmt.Errorf("failed to read status file '%s': %w", statusPath, err)
	}

	// Return true only if the server creator is "primary-server"
	if statusJson.ServerCreator == "primary-server" {
		return true, nil
	}

	return false, nil
}
