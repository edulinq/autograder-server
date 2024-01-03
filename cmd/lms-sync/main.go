package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/procedures"
    "github.com/eriq-augustine/autograder/util"
)

var args struct {
    config.ConfigArgs
    Course string `help:"ID of the course." arg:""`
    DryRun bool `help:"Do not actually do the operation, just state what you would do." default:"false"`
    SkipEmails bool `help:"Skip sending out emails to new users (always true if a dry run)." default:"false"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Sync course information from a course's LMS."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    db.MustOpen();
    defer db.MustClose();

    course := db.MustGetCourse(args.Course);

    result, err := procedures.SyncLMS(course, args.DryRun, !args.SkipEmails);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to sync LMS.");
    }

    if (result == nil) {
        fmt.Println("LMS sync not available for this course.");
        return;
    }

    fmt.Println(util.MustToJSONIndent(result));
}
