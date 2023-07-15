package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/model"
)

var args struct {
    Path []string `help:"Path to assignment JSON files." arg:"" type:"existingfile"`
}

func main() {
    kong.Parse(&args);

    for _, path := range args.Path {
        assignment := model.MustLoadAssignmentConfig(path);

        imageName, err := grader.BuildAssignmentImage(assignment);
        if (err != nil) {
            log.Fatal().Str("assignment", assignment.FullID()).Str("path", path).Err(err).Msg("Failed to build image.");
        }

        fmt.Printf("Built image '%s'.", imageName);
    }
}
