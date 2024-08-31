package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

const (
	API_REQUEST_CONTENT_KEY      = "content"
	API_SERVER_STOP_LOCK         = "API Lock"
	UNIX_SOCKET_SERVER_STOP_LOCK = "Unix Lock"
)

var (
	apiServer  *http.Server
	unixSocket net.Listener
)

// Run the API and Unix Socket Server.
func StartServers() error {
	errorsChan := make(chan error, 2)

	go func() {
		errorsChan <- runAPIServer()
	}()

	go func() {
		errorsChan <- runUnixSocketServer()
	}()

	// Handle server shutdowns for control c.
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-shutdownSignal
		signal.Stop(shutdownSignal)
		StopServers()
	}()

	// Wait for at least one error (or nil) to stop both servers,
	// then wait for the next error (or nil).
	var errs error = nil
	errs = errors.Join(errs, <-errorsChan)
	StopServers()
	errs = errors.Join(errs, <-errorsChan)

	close(errorsChan)

	return errs
}

func runAPIServer() (err error) {
	defer func() {
		value := recover()
		if value == nil {
			return
		}

		err = fmt.Errorf("API server panicked: '%v'.", value)
	}()

	var port = config.WEB_PORT.Get()

	log.Info("API Server Started", log.NewAttr("port", port))

	apiServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: core.GetRouteServer(GetRoutes()),
	}

	err = apiServer.ListenAndServe()
	if err == http.ErrServerClosed {
		err = nil
	}

	log.Info("API Server Stopped", log.NewAttr("port", port))

	if err != nil {
		log.Error("API server returned an error.", err)
	}

	return err
}

func StopServers() {
	StopUnixSocketServer()
	StopAPIServer()
}

func StopUnixSocketServer() {
	common.Lock(UNIX_SOCKET_SERVER_STOP_LOCK)
	defer common.Unlock(UNIX_SOCKET_SERVER_STOP_LOCK)

	if unixSocket == nil {
		return
	}

	tempUnixSocket := unixSocket
	unixSocket = nil

	err := tempUnixSocket.Close()
	if err != nil {
		log.Fatal("Failed to close the unix socket.", err)
	}

	err = util.RemoveDirent(common.GetStatusPath())
	if err != nil {
		log.Fatal("Failed to remove the status file.", err)
	}
}

func StopAPIServer() {
	common.Lock(API_SERVER_STOP_LOCK)
	defer common.Unlock(API_SERVER_STOP_LOCK)

	if apiServer == nil {
		return
	}

	tempApiServer := apiServer
	apiServer = nil

	err := tempApiServer.Shutdown(context.Background())
	if err != nil {
		log.Fatal("Failed to stop the API server.", err)
	}

	err = util.RemoveDirent(common.GetStatusPath())
	if err != nil {
		log.Fatal("Failed to remove the status file.", err)
	}
}
