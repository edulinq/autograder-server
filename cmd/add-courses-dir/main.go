package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
)

var args struct {
    config.ConfigArgs
    Path string `help:"Path to course JSON file." arg:"" type:"existingdir"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Add all specified courses to the system."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    db.MustOpen();
    defer db.MustClose();

    courseIDs, err := db.AddCoursesFromDir(args.Path, nil);
    if (err != nil) {
        log.Fatal().Err(err).Str("path", args.Path).Msg("Could not add courses dir.");
    }

    fmt.Printf("Added %d courses.\n", len(courseIDs));
    for _, courseID := range courseIDs {
        fmt.Printf("    %s\n", courseID);
    }
}
