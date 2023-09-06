package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
)

var args struct {
    config.ConfigArgs
}

func main() {
    kong.Parse(&args,
        kong.Description("Show all the known config options."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    jsonText, err := config.OptionsToJSON();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not serialize options.");
    }

    fmt.Println(jsonText);
}
