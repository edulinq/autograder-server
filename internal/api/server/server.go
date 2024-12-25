package server

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/edulinq/autograder/internal/api"
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/util"
)

// FinishCleanup ensures cleanup tasks execute before the server fully stops.
// Add(1) to the WaitGroup for each defered cleanup task and call Done() when finished.
var FinishCleanup sync.WaitGroup

// Run the autograder server and listen on an http and unix socket.
func RunServer(initiator common.ServerInitiator) (err error) {
	err = common.WriteAndHandleStatusFile(initiator)
	if err != nil {
		return err
	}

	apiDescription, err := api.Describe(*api.GetRoutes())
	if err != nil {
		return err
	}

	core.SetAPIDescription(*apiDescription)

	FinishCleanup.Add(1)
	defer func() {
		err = errors.Join(err, util.RemoveDirent(common.GetStatusPath()))
		FinishCleanup.Done()
	}()

	errorsChan := make(chan error, 2)

	go func() {
		errorsChan <- runAPIServer(api.GetRoutes())
	}()

	go func() {
		errorsChan <- runUnixSocketServer()
	}()

	// Gracefully shutdown on Control-C (SIGINT).
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-shutdownSignal
		signal.Stop(shutdownSignal)
		StopServer(true)
	}()

	// Wait for at least one error (or nil) to stop both servers,
	// then wait for the next error (or nil).
	err = errors.Join(err, <-errorsChan)
	// Stop server without waiting to ensure cleanup tasks get executed.
	StopServer(false)
	err = errors.Join(err, <-errorsChan)

	close(errorsChan)

	return err
}

// waitForCleanup should always be set to true when stopping a server
// to ensure the server is fully cleaned up by the time it's stopped.
func StopServer(waitForCleanup bool) {
	stopUnixSocketServer()
	stopAPIServer()

	if waitForCleanup {
		FinishCleanup.Wait()
	}
}
