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
        kong.Description("Dump all the loaded config and exit."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    fmt.Println("Basic Levels")
    log.Trace().Msg("This message is trace.");
    log.Debug().Msg("This message is debug.");
    log.Info().Msg("This message is info.");
    log.Warn().Msg("This message is warn.");
    log.Error().Msg("This message is error.");
    // log.Fatal().Msg("This message is fatal.");

    fmt.Println("Attatched Values");
    log.Info().Int("value", 1).Msg("Attatched Int.");
    log.Info().Float64("value", 2.3).Msg("Attatched Float64.");
    log.Info().Str("value", "Foo Bar").Msg("Attatched Str.");
    log.Info().Bool("value", true).Msg("Attatched Bool.");
    log.Info().Err(fmt.Errorf("This is an error!")).Msg("Attatched Err.");
}
