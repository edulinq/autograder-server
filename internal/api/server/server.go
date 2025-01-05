package server

import (
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/edulinq/autograder/internal/api"
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/common"
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
func (this *APIServer) RunAndBlock(initiator common.ServerInitiator) (err error) {
	err = common.WriteAndHandleStatusFile(initiator)
	if err != nil {
		return err
	}

	apiDescription, err := api.Describe(*api.GetRoutes())
	if err != nil {
		return err
	}

	core.SetAPIDescription(*apiDescription)

	defer func() {
		err = errors.Join(err, util.RemoveDirent(common.GetStatusPath()))
	}()

	go func() {
		this.errorsChan <- runAPIServer(api.GetRoutes())
	}()

	go func() {
		this.errorsChan <- runUnixSocketServer()
	}()

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

	close(this.errorsChan)

	return err
}

func (this *APIServer) Stop() {
	stopUnixSocketServer()
	stopAPIServer()
}
