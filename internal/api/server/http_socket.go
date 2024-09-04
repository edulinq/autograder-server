package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
)

var apiServer *http.Server

const API_SERVER_STOP_LOCK = "API_STOP_LOCK"

func runAPIServer(routes *[]*core.Route) (err error) {
	defer func() {
		value := recover()
		if value == nil {
			return
		}

		err = fmt.Errorf("API server panicked: '%v'.", value)
	}()

	var port = config.WEB_PORT.Get()

	log.Info("API Server Started.", log.NewAttr("port", port))

	apiServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: core.GetRouteServer(routes),
	}

	err = apiServer.ListenAndServe()
	if err == http.ErrServerClosed {
		// Set err to nil if the server stopped due to a graceful shutdown.
		err = nil
	}

	log.Info("API Server Stopped.", log.NewAttr("port", port))

	if err != nil {
		log.Error("API server returned an error.", err)
	}

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
		log.Fatal("Failed to stop the API server.", err)
	}
}
