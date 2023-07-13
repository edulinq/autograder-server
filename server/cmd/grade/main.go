package main

import (
    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/grader"
)

var args struct {
    Assignment string `help:"Path to assignment JSON files." required:"" type:"existingfile"`
    Submission string `help:"Path to submission directory." required:"" type:"existingdir"`
}

func main() {
    kong.Parse(&args);

    // TODO(eriq): Fetch info from config.

    err := grader.RunContainerGrader("autograder.test.p0", args.Submission);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to run container.");
    }
}
