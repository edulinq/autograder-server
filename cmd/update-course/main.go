package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
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
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    db.MustOpen();
    defer db.MustClose();

    course := db.MustGetCourse(args.Course);

    if (args.Clear) {
        err := db.ClearCourse(course);
        if (err != nil) {
            log.Fatal().Err(err).Msg("Failed to clear course.");
        }
    }

    if (args.Source != "") {
        course.Source = common.FileSpec(args.Source);
        err = db.SaveCourse(course);
        if (err != nil) {
            log.Fatal().Err(err).Msg("Failed to save course.");
        }
    }

    _, updated, err := db.UpdateCourseFromSource(course);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to update course.");
    }

    if (updated) {
        fmt.Println("Course updated.");
    } else {
        fmt.Println("No update available.")
    }
}
