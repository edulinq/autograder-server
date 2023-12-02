package main

import (
    "os"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/api"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/docker"
    "github.com/eriq-augustine/autograder/task"
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

    _, err = db.LoadCourses();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load courses.");
    }

    // Startup courses,

    for _, course := range db.MustGetCourses() {
        // Schedule tasks.
        for _, courseTask := range course.GetTasks() {
            err = task.Schedule(course, courseTask);
            if (err != nil) {
                log.Fatal().Err(err).Str("course-id", course.GetID()).Str("task", courseTask.String()).Msg("Failed to schedule task.");
            }
        }

        // Build images (in the background).
        go func() {
            _, errs := course.BuildAssignmentImages(false, false, docker.NewBuildOptions());
            for imageName, err := range errs {
                log.Error().Err(err).Str("course-id", course.GetID()).Str("image", imageName).Msg("Failed to build image.");
            }
        }();
    }

    err = api.StartServer();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Server was stopped.");
    }

    log.Info().Msg("Server closed.");
}
