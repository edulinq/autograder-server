package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/util"
)

var args struct {
    config.ConfigArgs
    Course string `help:"ID of the course." arg:""`
    Assignment string `help:"ID of the assignment." arg:""`
    Submission string `help:"Path to submission directory." required:"" type:"existingdir"`
    OutPath string `help:"Option path to output a JSON grading result." type:"path"`
    User string `help:"User email for the submission." default:"testuser"`
    Message string `help:"Submission message." default:""`
}

func main() {
    kong.Parse(&args);
    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    db.MustOpen();
    defer db.MustClose();

    assignment := db.MustGetAssignment(args.Course, args.Assignment);

    result, _, output, err := grader.GradeDefault(assignment, args.Submission, args.User, args.Message);
    if (err != nil) {
        if (output != "") {
            fmt.Println("Grading failed, but output was recovered:");
            fmt.Println(output);
        }
        log.Fatal().Err(err).Msg("Failed to run grader.");
    }

    if (args.OutPath != "") {
        err = util.ToJSONFileIndent(result, args.OutPath);
        if (err != nil) {
            log.Fatal().Err(err).Str("outpath", args.OutPath).Msg("Failed to output JSON result.");
        }
    }

    fmt.Println(result.Report());
}
