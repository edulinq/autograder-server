package main

import (
    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/lms/lmsusers"
)

var args struct {
    config.ConfigArgs
    Course string `help:"ID of the course." arg:""`
    DryRun bool `help:"Do not actually do the operation, just state what you would do." default:"false"`
    SkipEmails bool `help:"Skip sending out emails to new users (always true if a dry run)." default:"false"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Sync local users with matching LMS users."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    db.MustOpen();
    defer db.MustClose();

    course := db.MustGetCourse(args.Course);

    result, err := lmsusers.SyncLMSUsers(course, args.DryRun, !args.SkipEmails);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to sync LMS users.");
    }

    result.PrintReport();
}
