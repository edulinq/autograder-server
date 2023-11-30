package api

// Run the standard API server.

import (
    "fmt"
    "net/http"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/config"
)

func StartServer() error {
    var port = config.WEB_PORT.Get();

    log.Info().Msgf("Serving on %d.", port);
    return http.ListenAndServe(fmt.Sprintf(":%d", port), core.GetRouteServer(GetRoutes()));
}
