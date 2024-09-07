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
func StartServer() (err error) {
	err = common.WriteAndHandleStatusFile()
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, util.RemoveDirent(common.GetStatusPath()))
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
		StopServer()
	}()

	// Wait for at least one error (or nil) to stop both servers,
	// then wait for the next error (or nil).
	err = errors.Join(err, <-errorsChan)
	StopServer()
	err = errors.Join(err, <-errorsChan)

	close(errorsChan)

	return err
}

func StopServer() {
	stopUnixSocketServer()
	stopAPIServer()
}
