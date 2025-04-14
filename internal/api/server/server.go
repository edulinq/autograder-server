package server

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/edulinq/autograder/internal/api"
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/systemserver"
	"github.com/edulinq/autograder/internal/util"
)

type APIServer struct {
	errorsChan     chan error
	shutdownSignal chan os.Signal
}

func NewAPIServer() *APIServer {
	return &APIServer{
		errorsChan:     make(chan error),
		shutdownSignal: make(chan os.Signal),
	}
}

// Run the autograder server and listen on an http and unix socket.
func (this *APIServer) RunAndBlock(initiator systemserver.ServerInitiator) (err error) {
	err = systemserver.WriteAndHandleStatusFile(initiator)
	if err != nil {
		return err
	}

	core.SetAPIRoutes(api.GetRoutes())

	defer func() {
		err = errors.Join(err, util.RemoveDirent(systemserver.GetStatusPath()))
	}()

	// Create a wait group for the respective servers to indicate that they are now waiting.
	// We do this so we don't try to stop a server before it has started
	// (e.g., if the other server failed on setup).
	var subserverSetupWaitGroup sync.WaitGroup
	subserverSetupWaitGroup.Add(2)

	go func() {
		this.errorsChan <- runAPIServer(api.GetRoutes(), &subserverSetupWaitGroup)
	}()

	go func() {
		this.errorsChan <- runUnixSocketServer(&subserverSetupWaitGroup)
	}()

	// Wait for the subservers to be ready before trying to stop them.
	subserverSetupWaitGroup.Wait()

	// Gracefully shutdown on Control-C (SIGINT).
	signal.Notify(this.shutdownSignal, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-this.shutdownSignal
		signal.Stop(this.shutdownSignal)
		this.Stop()
	}()

	// Wait for at least one error (or nil) to stop both servers,
	// then wait for the next error (or nil).
	err = errors.Join(err, <-this.errorsChan)

	// Stop server without waiting to ensure cleanup tasks get executed.
	this.Stop()
	err = errors.Join(err, <-this.errorsChan)

	return err
}

func (this *APIServer) Stop() {
	stopUnixSocketServer()
	stopAPIServer()
}
