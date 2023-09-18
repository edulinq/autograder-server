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
    grader.DockerBuildOptions
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

    var imageNames []string;

    if (len(args.Path) > 0) {
        imageNames = buildFromPaths(args.Path, &args.DockerBuildOptions);
    } else {
        imageNames = buildFromCourses(&args.DockerBuildOptions);
    }

    fmt.Printf("Successfully built %d images:\n", len(imageNames));
    for _, imageName := range imageNames {
        fmt.Printf("    %s\n", imageName);
    }
}

func buildFromCourses(buildOptions *grader.DockerBuildOptions) []string {
    err := grader.LoadCourses();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to load courses.");
    }

    imageNames, errs := grader.BuildDockerImages(buildOptions);
    if (len(errs) > 0) {
        for _, err = range errs {
            log.Error().Err(err).Msg("Failed to build grader docker images.");
        }

        log.Fatal().Int("count", len(errs)).Msg("Failed to build course images.");
    }

    return imageNames;
}

func buildFromPaths(paths []string, buildOptions *grader.DockerBuildOptions) []string {
    imageNames := make([]string, 0);

    for _, path := range paths {
        assignment := model.MustLoadAssignmentConfig(path);

        err := grader.BuildDockerImageWithOptions(assignment, buildOptions);
        if (err != nil) {
            log.Fatal().Str("assignment", assignment.FullID()).Str("path", path).Err(err).Msg("Failed to build image.");
        }

        imageNames = append(imageNames, assignment.ImageName());
    }

    return imageNames;
}
