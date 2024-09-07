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

const API_SERVER_STOP_LOCK = "internal.api.server.API_STOP_LOCK"

func runAPIServer(routes *[]*core.Route) (err error) {
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
		err = nil // Set err to nil if the API server stopped due to a graceful shutdown.
	}

	log.Info("API Server Stopped.", log.NewAttr("port", port))

	if err != nil {
		log.Error("API server returned an error.", err)
	}

	return err
}

func stopAPIServer() {
	common.Lock(API_SERVER_STOP_LOCK)
	defer func() {
		err := common.Unlock(API_SERVER_STOP_LOCK)
		if err != nil {
			log.Error("Failed to unlock the api server stop lock: '%w'.", err)
		}
	}()

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
