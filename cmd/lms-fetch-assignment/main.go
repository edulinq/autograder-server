package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/lms"
    "github.com/eriq-augustine/autograder/util"
)

var args struct {
    config.ConfigArgs
    Course string `help:"ID of the course." arg:""`
    Assignment string `help:"ID of the assignment." arg:""`
}

func main() {
    kong.Parse(&args,
        kong.Description("Fetch the grades for a specific assignment from an LMS." +
            " Either --assignment-path or (--course-path and --assignment-id) are required."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    db.MustOpen();
    defer db.MustClose();

    assignment := db.MustGetAssignment(args.Course, args.Assignment);
    course := assignment.GetCourse();

    if (assignment.GetLMSID() == "") {
        log.Fatal().Str("assignment", assignment.FullID()).Msg("Assignment has no LMS ID.");
    }

    lmsAssignment, err := lms.FetchAssignment(course, assignment.GetLMSID());
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not fetch assignment.");
    }

    fmt.Println(util.MustToJSONIndent(lmsAssignment));
}
