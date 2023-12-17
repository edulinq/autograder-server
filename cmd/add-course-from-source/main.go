package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/util"
)

var args struct {
    config.ConfigArgs
    Source string `help:"The source to add a course from." arg:""`
}

func main() {
    kong.Parse(&args,
        kong.Description("Add a course to system from a source (FileSpec)."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    db.MustOpen();
    defer db.MustClose();

    spec, err := common.ParseFileSpec(args.Source);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to parse FileSpec.");
    }

    tempDir, err := util.MkDirTemp("autograder-add-course-source-");
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to make temp source dir.");
    }
    defer util.RemoveDirent(tempDir);

    err = spec.CopyTarget(common.ShouldGetCWD(), tempDir, false);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to copy source.");
    }

    courseIDs, err := db.AddCoursesFromDir(tempDir, spec);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to add course dir.");
    }

    fmt.Printf("Added %d courses.\n", len(courseIDs));
    for _, courseID := range courseIDs {
        fmt.Printf("    %s\n", courseID);
    }
}
