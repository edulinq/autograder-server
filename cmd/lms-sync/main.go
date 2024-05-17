package main

import (
    "fmt"

    "github.com/alecthomas/kong"

    "github.com/edulinq/autograder/internal/config"
    "github.com/edulinq/autograder/internal/db"
    "github.com/edulinq/autograder/internal/lms/lmssync"
    "github.com/edulinq/autograder/internal/log"
    "github.com/edulinq/autograder/internal/util"
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
        log.Fatal("Could not load config options.", err);
    }

    db.MustOpen();
    defer db.MustClose();

    course := db.MustGetCourse(args.Course);

    result, err := lmssync.SyncLMS(course, args.DryRun, !args.SkipEmails);
    if (err != nil) {
        log.Fatal("Failed to sync LMS.", err, course);
    }

    if (result == nil) {
        fmt.Println("LMS sync not available for this course.");
        return;
    }

    fmt.Println(util.MustToJSONIndent(result));
}
