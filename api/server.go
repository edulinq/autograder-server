package api

import (
    "fmt"
    "net/http"

    "github.com/edulinq/autograder/api/core"
    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/log"
)

// Run the standard API server.
func StartServer() error {
    var port = config.WEB_PORT.Get();

    log.Info("API Server Started", log.NewAttr("port", port));
    return http.ListenAndServe(fmt.Sprintf(":%d", port), core.GetRouteServer(GetRoutes()));
}
