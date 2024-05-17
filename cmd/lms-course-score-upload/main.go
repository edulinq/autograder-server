package main

import (
    "fmt"

    "github.com/alecthomas/kong"

    "github.com/edulinq/autograder/internal/config"
    "github.com/edulinq/autograder/internal/db"
    "github.com/edulinq/autograder/internal/log"
    "github.com/edulinq/autograder/internal/scoring"
)

var args struct {
    config.ConfigArgs
    Course string `help:"ID of the course." arg:""`
    DryRun bool `help:"Do not actually upload the grades, just state what you would do." default:"false"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Perform a full course scoring (including late policy) and upload."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal("Could not load config options.", err);
    }

    db.MustOpen();
    defer db.MustClose();

    course := db.MustGetCourse(args.Course);

    err = scoring.FullCourseScoringAndUpload(course, args.DryRun);
    if (err != nil) {
        log.Fatal("Failed to score and upload assignment.", err, course);
    }

    fmt.Println("Course grades uploaded.");
}
