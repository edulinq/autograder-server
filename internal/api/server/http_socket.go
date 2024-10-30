package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
)

var apiServer *http.Server

const API_SERVER_STOP_LOCK = "internal.api.server.API_SERVER_STOP_LOCK"
const PORT_USAGE_CHECK_LOCK = "internal.api.server.PORT_USAGE_CHECK_LOCK"

func runAPIServer(routes *[]*core.Route) (err error) {
	var port = config.WEB_PORT.Get()

	common.Lock(PORT_USAGE_CHECK_LOCK)
	err = checkPortInUse(port)
	if err != nil {
		return err
	}
	common.Unlock(PORT_USAGE_CHECK_LOCK)

	defer func() {
		value := recover()
		if value == nil {
			return
		}

		err = errors.Join(err, fmt.Errorf("API server panicked: '%v'.", value))
	}()

	log.Info("API Server Started.", log.NewAttr("port", port))

	apiServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: core.GetRouteServer(routes),
	}

	err = apiServer.ListenAndServe()
	if err == http.ErrServerClosed {
		// Set err to nil if the API server stopped due to a graceful shutdown.
		err = nil
	}

	if err != nil {
		log.Error("API server returned an error.", err)
	}

	log.Info("API Server Stopped.", log.NewAttr("port", port))

	return err
}

func stopAPIServer() {
	common.Lock(API_SERVER_STOP_LOCK)
	defer common.Unlock(API_SERVER_STOP_LOCK)

	if apiServer == nil {
		return
	}

	tempApiServer := apiServer
	apiServer = nil

	err := tempApiServer.Shutdown(context.Background())
	if err != nil {
		log.Error("Failed to stop the API server.", err)
	}
}

func checkPortInUse(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("Port '%d' is already in use or not available: '%w'.", port, err)
	}

	err = listener.Close()
	if err != nil {
		return fmt.Errorf("Failed to close the listener after checking port usage: '%w'.", err)
	}

	return nil
}
