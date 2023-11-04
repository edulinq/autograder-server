package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/task"
)

var args struct {
    config.ConfigArgs
    Course string `help:"ID of the course." arg:"" optional:""`
}

func main() {
    kong.Parse(&args,
        kong.Description("Backup a single course or all known courses."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    var courses map[string]model.Course;

    if (args.Course != "") {
        course := db.MustGetCourse(args.Course);
        courses[course.GetID()] = course;
    } else {
        courses = db.MustGetCourses();
    }

    courseIDs := backupFromMap(courses);

    fmt.Printf("Successfully backed-up %d courses:\n", len(courseIDs));
    for _, courseID := range courseIDs {
        fmt.Printf("    %s\n", courseID);
    }
}

func backupFromMap(courses map[string]model.Course) []string {
    courseIDs := make([]string, 0);
    errs := make([]error, 0);

    for _, course := range courses {
        err := task.RunBackup(course.GetSourceDir(), "", course.GetID());
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
