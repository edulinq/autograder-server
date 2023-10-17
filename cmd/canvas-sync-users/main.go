package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
)

var args struct {
    config.ConfigArgs
    Path string `help:"Path to course JSON file." arg:"" type:"existingfile"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Sync IDs with matching LMS users (does not add/remove users)."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    course := model.MustLoadCourseConfig(args.Path);
    if (course.LMSAdapter == nil) {
        log.Fatal().Msg("Course has no LMS info associated with it.");
    }

    count, err := course.SyncLMSUsers();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to sync LMS users.");
    }

    fmt.Printf("Updated %d users.\n", count);
}
