package main

import (
    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/model"
)

var args struct {
    Assignment string `help:"Path to assignment JSON files." required:"" type:"existingfile"`
    Submission string `help:"Path to submission directory." required:"" type:"existingdir"`
}

func main() {
    kong.Parse(&args);

    assignment := model.MustLoadAssignmentConfig(args.Assignment);

    err := grader.RunContainerGrader(assignment.ImageName(), args.Submission);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to run container.");
    }
}
