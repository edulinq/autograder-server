package main

import (
    "os"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/util"
    "github.com/eriq-augustine/autograder/web"
)

var args struct {
    config.ConfigArgs
}

func main() {
    kong.Parse(&args);
    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    log.Info().Str("version", util.GetAutograderFullVersion()).Msg("Autograder Version");

    workingDir, err := os.Getwd();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not get working directory.");
    }

    db.MustOpen();
    defer db.MustClose();

    log.Info().Str("dir", workingDir).Msg("Running server with working directory.");

    _, err = db.LoadCourses();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load courses.");
    }

    /* TEST
    err = grader.ActivateCourses();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not activate courses.");
    }
    */

    web.StartServer();
}
