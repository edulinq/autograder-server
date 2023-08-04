package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/model"
)

var args struct {
    config.ConfigArgs
    Path []string `help:"Path to assignment JSON files." arg:"" optional:"" type:"existingfile"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Build all images from all known assignments (if no paths are supplied), or the images specified by the given assignments."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    if (len(args.Path) > 0) {
        buildFromPaths(args.Path);
        return;
    }

    buildFromCourses();
}

func buildFromCourses() {
    err := grader.LoadCourses();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to load courses.");
    }

    courses := grader.GetCourses();

    for _, course := range courses {
        for _, assignment := range course.Assignments {
            err = assignment.BuildDockerImage();
            if (err != nil) {
                log.Fatal().Str("assignment", assignment.FullID()).Str("course", course.ID).Err(err).Msg("Failed to build image.");
            }
        }
    }
}

func buildFromPaths(paths []string) {
    for _, path := range paths {
        assignment := model.MustLoadAssignmentConfig(path);

        err := assignment.BuildDockerImage();
        if (err != nil) {
            log.Fatal().Str("assignment", assignment.FullID()).Str("path", path).Err(err).Msg("Failed to build image.");
        }

        fmt.Printf("Built image '%s' from path '%s'.", assignment.ImageName(), path);
    }
}
