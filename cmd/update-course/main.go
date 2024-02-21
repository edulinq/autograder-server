package main

import (
    "fmt"

    "github.com/alecthomas/kong"

    "github.com/edulinq/autograder/common"
    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/db"
    "github.com/edulinq/autograder/log"
    "github.com/edulinq/autograder/procedures"
)

var args struct {
    config.ConfigArgs
    Course string `help:"ID of the course." arg:""`
    Source string `help:"An optional new source for the course."`
    Clear bool `help:"Clear the course before updating." default:"false"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Update a course with the existing (or new) source."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal("Could not load config options.", err);
    }

    db.MustOpen();
    defer db.MustClose();

    course := db.MustGetCourse(args.Course);

    if (args.Clear) {
        err := db.ClearCourse(course);
        if (err != nil) {
            log.Fatal("Failed to clear course.", err, course);
        }
    }

    if (args.Source != "") {
        spec, err := common.ParseFileSpec(args.Source);
        if (err != nil) {
            log.Fatal("Failed to parse FileSpec.", err, course);
        }

        course.Source = spec;

        err = db.SaveCourse(course);
        if (err != nil) {
            log.Fatal("Failed to save course.", err, course);
        }
    }

    updated, err := procedures.UpdateCourse(course, false);
    if (err != nil) {
        log.Fatal("Failed to update course.", err, course);
    }

    if (updated) {
        fmt.Println("Course updated.");
    } else {
        fmt.Println("No update available.")
    }
}
