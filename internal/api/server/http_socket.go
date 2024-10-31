package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
)

var apiServer *http.Server

const API_SERVER_LOCK = "internal.api.server.API_SERVER_LOCK"

func runAPIServer(routes *[]core.Route) (err error) {
	common.Lock(API_SERVER_LOCK)
	if apiServer != nil {
		common.Unlock(API_SERVER_LOCK)
		return fmt.Errorf("API server is already running.")
	}
	common.Unlock(API_SERVER_LOCK)

	defer func() {
		value := recover()
		if value == nil {
			return
		}

		err = errors.Join(err, fmt.Errorf("API server panicked: '%v'.", value))
	}()

	var port = config.WEB_PORT.Get()

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
	common.Lock(API_SERVER_LOCK)
	defer common.Unlock(API_SERVER_LOCK)

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
