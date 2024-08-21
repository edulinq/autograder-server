package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
)

const API_REQUEST_CONTENT_KEY = "content"

var (
	apiServer  *http.Server
	unixServer *http.Server
)

func StartServer() error {
	var serverShutdown sync.WaitGroup
	serverError := make(chan error, 2)

	serverShutdown.Add(2)
	go func() {
		defer serverShutdown.Done()

		err := startAPIServer()
		if err != nil {
			log.Error("Failed to start the API server.", err)
			serverError <- err
		}
	}()

	go func() {
		defer serverShutdown.Done()

		err := startUnixSocketServer()
		if err != nil {
			log.Error("Failed to start the Unix Socket server.", err)
			serverError <- err
		}
	}()

	go func() {
		serverShutdown.Wait()
		close(serverError)
	}()

	select {
	case err := <-serverError:
		StopServers()
		return err
	case <-serverError:
		StopServers()
	}

	return nil
}

func startAPIServer() error {
	var port = config.WEB_PORT.Get()

	log.Info("API Server Started", log.NewAttr("port", port))
	apiServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: core.GetRouteServer(GetRoutes()),
	}

	return apiServer.ListenAndServe()
}

func StopServers() {
	err := apiServer.Shutdown(context.Background())
	if err != nil {
		log.Fatal("Failed to stop the API server.", err)
	}

	err = os.Remove(config.PID_PATH.Get())
	if err != nil {
		log.Fatal("Failed to remove the PID file.", err)
	}

	if unixSocket != nil {
		err := unixSocket.Close()
		if err != nil {
			log.Fatal("Failed to close the unix socket.", err)
		}

		err = os.Remove(config.UNIX_SOCKET_PATH.Get())
		if err != nil {
			log.Fatal("Failed to remove the unix socket file.", err)
		}
	}
}
