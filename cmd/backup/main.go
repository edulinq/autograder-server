package main

import (
    "fmt"

    "github.com/alecthomas/kong"

    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/db"
    "github.com/edulinq/autograder/log"
    "github.com/edulinq/autograder/model"
    "github.com/edulinq/autograder/task"
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
        log.Fatal("Could not load config options.", err);
    }

    db.MustOpen();
    defer db.MustClose();

    var courses map[string]*model.Course;

    if (args.Course != "") {
        courses = make(map[string]*model.Course);
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

func backupFromMap(courses map[string]*model.Course) []string {
    courseIDs := make([]string, 0);
    errorCount := 0;

    for _, course := range courses {
        err := task.RunBackup(course, "", "");
        if (err != nil) {
            log.Error("Failed to backup course.", err, course);
            errorCount++;
        } else {
            courseIDs = append(courseIDs, course.GetID());
        }
    }

    if (errorCount > 0) {
        log.Fatal("Failed to backup courses.", log.NewAttr("error-count", errorCount));
    }

    return courseIDs;
}
