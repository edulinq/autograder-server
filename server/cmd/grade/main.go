package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

var args struct {
    config.ConfigArgs
    Assignment string `help:"Path to assignment JSON files." required:"" type:"existingfile"`
    Submission string `help:"Path to submission directory." required:"" type:"existingdir"`
    OutPath string `help:"Option path to output a JSON grading result." type:"path"`
    User string `help:"User email for the submission." default:"testuser"`
    NoStore bool `help:"Do not store the grading result into the user's submission directory." default:"false"`
    Debug bool `help:"Leave some debug artifacts like the grading sirectory." default:"false"`
}

func main() {
    kong.Parse(&args);
    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    assignment := model.MustLoadAssignmentConfig(args.Assignment);

    options := grader.GetDefaultGradeOptions();
    options.LeaveTempDir = args.Debug;
    options.UseFakeSubmissionsDir = args.NoStore;

    result, err := grader.Grade(assignment, args.Submission, args.User, options);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to run grader.");
    }

    if (args.OutPath != "") {
        err = util.ToJSONFileIndent(result, args.OutPath, "", "    ");
        if (err != nil) {
            log.Fatal().Err(err).Str("outpath", args.OutPath).Msg("Failed to output JSON result.");
        }
    }

    fmt.Println(result.Report());
}
