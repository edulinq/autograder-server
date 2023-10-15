package api

// Run the standard API server.

import (
    "fmt"
    "net/http"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/config"
)

func StartServer() {
    var port = config.WEB_PORT.GetInt();

    log.Info().Msgf("Serving on %d.", port);

    err := http.ListenAndServe(fmt.Sprintf(":%d", port), core.GetRouteServer(GetRoutes()));
    if (err != nil) {
        log.Fatal().Err(err).Msg("Server stopped.");
    }
}
