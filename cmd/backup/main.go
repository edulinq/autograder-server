package main

import (
    "fmt"
    "path/filepath"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/task"
)

var args struct {
    config.ConfigArgs
    Path []string `help:"Path to course JSON files." arg:"" optional:"" type:"existingfile"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Backup all known courses (if no paths are supplied), or just the courses specified."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    var courseIDs []string;

    if (len(args.Path) > 0) {
        courseIDs = backupFromPaths(args.Path);
    } else {
        courseIDs = backupFromCourses();
    }

    fmt.Printf("Successfully backed-up %d courses:\n", len(courseIDs));
    for _, courseID := range courseIDs {
        fmt.Printf("    %s\n", courseID);
    }
}

func backupFromCourses() []string {
    err := grader.LoadCourses();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to load courses.");
    }

    return backupFromMap(grader.GetCourses());
}

func backupFromPaths(paths []string) []string {
    courses := make(map[string]*model.Course);

    for _, path := range paths {
        course := model.MustLoadCourseConfig(path);
        courses[course.GetID()] = course;
    }

    return backupFromMap(courses);
}

func backupFromMap(courses map[string]*model.Course) []string {
    courseIDs := make([]string, 0);
    errs := make([]error, 0);

    for _, course := range courses {
        err := task.RunBackup(filepath.Dir(course.SourcePath), "", course.GetID());
        if (err != nil) {
            errs = append(errs, fmt.Errorf("Failed to backup course '%s': '%w'.", course.GetID(), err));
        } else {
            courseIDs = append(courseIDs, course.GetID());
        }
    }

    if (len(errs) > 0) {
        for _, err := range errs {
            log.Error().Err(err).Msg("Failed to backup course.");
        }

        log.Fatal().Int("count", len(errs)).Msg("Failed to backup courses.");
    }

    return courseIDs;
}
