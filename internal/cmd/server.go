package cmd

import (
	"sync"
	"time"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/procedures/server"
	"github.com/edulinq/autograder/internal/systemserver"
	"github.com/edulinq/autograder/internal/util"
)

// Check to see if a server is running and start a CMD server if it's not.
// Returns nil if the primary server or CMD test server is already running,
// a cleanup function if the CMD started its own server,
// or exits (log.Fatal()) if another CMD server is already running.
// If the cleanup function is not nil, it should be called after the caller is done with the server.
func mustEnsureServerIsRunning() func() {
	statusInfo, err := systemserver.CheckAndHandleServerStatusFile()
	if err != nil {
		log.Fatal("Failed to retrieve the current status file's json.", err)
	}

	if statusInfo != nil {
		switch statusInfo.Initiator {
		// log.Fatal() if another CMD server is running since they share the same working directory.
		case systemserver.CMD_SERVER:
			log.Fatal("Cannot start server, another CMD server is running.", log.NewAttr("PID", statusInfo.Pid))
		// Don't start the CMD server if the primary server or CMD test server is running.
		default:
			return nil
		}
	}

	port, err := util.GetUnusedPort()
	if err != nil {
		log.Fatal("Failed to get an unused port.", err)
	}

	// Store old config values so we can restore them later.
	oldHTTPS := config.WEB_HTTPS_ENABLE.Get()
	oldPort := config.WEB_HTTP_PORT.Get()

	// Set the config so we are just running HTTP on a random unused port.
	config.WEB_HTTPS_ENABLE.Set(false)
	config.WEB_HTTP_PORT.Set(port)

	var serverStart sync.WaitGroup
	serverStart.Add(1)

	go func() {
		serverStart.Done()

		err = server.RunAndBlock(systemserver.CMD_SERVER)
		if err != nil {
			log.Fatal("Failed to start the server.", err)
		}
	}()

	serverStart.Wait()

	// Small sleep to allow the server to start up.
	time.Sleep(150 * time.Millisecond)

	// Create a cleanup function that resets config options and stops the sevrer.
	cleanupFunc := func() {
		config.WEB_HTTP_PORT.Set(oldPort)
		config.WEB_HTTPS_ENABLE.Set(oldHTTPS)

		err := server.CleanupAndStop()
		if err != nil {
			log.Fatal("Failed to cleanup and stop the CMD server.", err)
		}
	}

	return cleanupFunc
}
