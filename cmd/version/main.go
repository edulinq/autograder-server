package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/util"
)

var args struct {
    config.ConfigArgs
}

func main() {
    kong.Parse(&args,
        kong.Description("Get the autograder's version."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    fmt.Println("Autograder");
    fmt.Printf("Short Version: %s\n", util.GetAutograderVersion());
    fmt.Printf("Full  Version: %s\n", util.GetAutograderFullVersion());
    fmt.Printf("API   Version: %d\n", core.API_VERSION);
}
