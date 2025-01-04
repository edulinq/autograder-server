package cmd

import (
	"sync"
	"time"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/procedures/server"
	"github.com/edulinq/autograder/internal/util"
)

// Check to see if a server is running and start a CMD server if it's not.
// Returns (false, 0) if the primary server or CMD test server is already running,
// (true, oldPort) if the CMD started its own server,
// or log.Fatal() if another CMD server is already running.
func mustEnsureServerIsRunning() (bool, int) {
	statusInfo, err := common.CheckAndHandleServerStatusFile()
	if err != nil {
		log.Fatal("Failed to retrieve the current status file's json.", err)
	}

	if statusInfo != nil {
		switch statusInfo.Initiator {
		// log.Fatal() if another CMD server is running since they share the same working directory.
		case common.CMD_SERVER:
			log.Fatal("Cannot start server, another CMD server is running.", log.NewAttr("PID", statusInfo.Pid))
		// Don't start the CMD server if the primary server or CMD test server is running.
		default:
			return false, 0
		}
	}

	port, err := util.GetUnusedPort()
	if err != nil {
		log.Fatal("Failed to get an unused port.", err)
	}

	oldPort := config.WEB_PORT.Get()
	config.WEB_PORT.Set(port)

	var serverStart sync.WaitGroup
	serverStart.Add(1)

	go func() {
		serverStart.Done()

		err = server.RunAndBlock(common.CMD_SERVER)
		if err != nil {
			log.Fatal("Failed to start the server.", err)
		}
	}()

	serverStart.Wait()

	// Small sleep to allow the server to start up.
	time.Sleep(100 * time.Millisecond)

	return true, oldPort
}
