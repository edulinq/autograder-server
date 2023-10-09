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
    Path []string `help:"Path to course JSON file." arg:"" optional:"" type:"existingfile"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Load all courses with the specified paths, or the default courses from config if no paths are specified."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    count := 0;

    for _, path := range args.Path {
        course := model.MustLoadCourseConfig(path);
        count++;
        fmt.Printf("Loaded course '%s'.\n", course.ID);
    }

    if (count == 0) {
        err = grader.LoadCourses();
        if (err != nil) {
            log.Fatal().Err(err).Msg("Could not load courses.");
        }

        for _, course := range grader.GetCourses() {
            fmt.Printf("Loaded course '%s'.\n", course.ID);
            count++;
        }
    }

    fmt.Printf("Loaded %d courses.\n", count);
}
