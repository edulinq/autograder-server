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
        // TODO(eriq): Find course config file.
        courseName := "test";

        config, err := model.LoadAssignmentConfig(path);
        if (err != nil) {
            log.Fatal().Str("course", courseName).Str("path", path).Err(err).Msg("Failed to load assignment config.");
        }

        imageName, err := grader.BuildAssignmentImage(courseName, config);
        if (err != nil) {
            log.Fatal().Str("course", courseName).Str("path", path).Err(err).Msg("Failed to build image.");
        }

        fmt.Printf("Built image '%s'.", imageName);
    }
}
