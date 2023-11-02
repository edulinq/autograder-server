package main

import (
    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
)

var args struct {
    config.ConfigArgs
    Path string `help:"Path to course JSON file." arg:"" type:"existingfile"`
    DryRun bool `help:"Do not actually do the operation, just state what you would do." default:"false"`
    SkipSendEmails bool `help:"Skip sending out emails to new users (always true if a dry run)." default:"false"`
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
    if (course.GetLMSAdapter() == nil) {
        log.Fatal().Msg("Course has no LMS info associated with it.");
    }

    args.SkipSendEmails = (args.SkipSendEmails || args.DryRun);
    result, err := course.SyncLMSUsers(args.DryRun, !args.SkipSendEmails);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to sync LMS users.");
    }

    result.PrintReport();
}
