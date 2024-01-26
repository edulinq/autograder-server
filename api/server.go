package api

import (
    "fmt"
    "net/http"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/log"
)

// Run the standard API server.
func StartServer() error {
    var port = config.WEB_PORT.Get();

    log.Info("API Server Started", log.NewAttr("port", port));
    return http.ListenAndServe(fmt.Sprintf(":%d", port), core.GetRouteServer(GetRoutes()));
}
