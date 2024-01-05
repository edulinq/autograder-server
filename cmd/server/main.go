package main

import (
    "fmt"
    "os"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/api"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/procedures"
    "github.com/eriq-augustine/autograder/util"
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

    _, err = db.AddCourses();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load courses.");
    }

    courses := db.MustGetCourses();
    log.Info().Int("count", len(courses)).Msg(fmt.Sprintf("Loaded %d course(s).", len(courses)));

    // Startup courses (in the background).
    for _, course := range courses {
        log.Info().Str("course", course.GetID()).Msg("Loaded course.");
        go func(course *model.Course) {
            procedures.UpdateCourse(course, true);
        }(course);
    }

    // Cleanup any temp dirs.
    defer util.RemoveRecordedTempDirs();

    err = api.StartServer();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Server was stopped.");
    }

    log.Info().Msg("Server closed.");
}
