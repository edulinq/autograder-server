package api

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
)

var API_REQUEST_CONTENT_KEY = "content"

func StartServer() error {
	var serverShutdown sync.WaitGroup
	serverShutdown.Add(1)

	go func() {
		defer serverShutdown.Done()

		err := startExclusiveAPIServer()
		if err != nil {
			log.Error("Failed to start the api server.", err)
		}
	}()

	go func() {
		defer serverShutdown.Done()

		err := startExclusiveUnixServer()
		if err != nil {
			log.Error("Failed to start the unix server.", err)
		}
	}()

	serverShutdown.Wait()

	return nil
}

func startExclusiveAPIServer() error {
	var port = config.WEB_PORT.Get()

	log.Info("API Server Started", log.NewAttr("port", port))
	return http.ListenAndServe(fmt.Sprintf(":%d", port), core.GetRouteServer(GetRoutes()))
}
