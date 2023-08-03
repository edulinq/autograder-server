package main

import (
    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
)

var args struct {
    config.ConfigArgs
    Assignment string `help:"Path to assignment JSON files." required:"" type:"existingfile"`
    Submission string `help:"Path to submission directory." required:"" type:"existingdir"`
}

func main() {
    kong.Parse(&args);
    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    assignment := model.MustLoadAssignmentConfig(args.Assignment);

    err = assignment.RunGrader(args.Submission);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to run container.");
    }
}
