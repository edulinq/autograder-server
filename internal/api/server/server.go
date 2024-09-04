package server

import (
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/edulinq/autograder/internal/api"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/util"
)

// Run the autograder server and listen on an http and unix socket.
func StartServer() error {
	err := common.WriteAndHandleStatusFile()
	if err != nil {
		return err
	}
	defer util.RemoveDirent(common.GetStatusPath())

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
		StopServer()
	}()

	// Wait for at least one error (or nil) to stop both servers,
	// then wait for the next error (or nil).
	var errs error = nil
	errs = errors.Join(errs, <-errorsChan)
	StopServer()
	errs = errors.Join(errs, <-errorsChan)

	close(errorsChan)

	return errs
}

func StopServer() {
	stopUnixSocketServer()
	stopAPIServer()
}
