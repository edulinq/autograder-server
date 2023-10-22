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
    DryRun bool `help:"Do not actually upload the grades, just state what you would do." default:"false"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Perform a full course scoring (including late policy) and upload."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    course := model.MustLoadCourseConfig(args.Path);
    if (course.LMSAdapter == nil) {
        log.Fatal().Msg("Course has no LMS info associated with it.");
    }

    err = course.FullScoringAndUpload(args.DryRun);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to score and upload assignment.");
    }

    fmt.Println("Course grades uploaded.");
}
