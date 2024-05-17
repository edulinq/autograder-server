package main

import (
    "fmt"

    "github.com/alecthomas/kong"

    "github.com/edulinq/autograder/internal/config"
    "github.com/edulinq/autograder/internal/db"
    "github.com/edulinq/autograder/internal/log"
)

var args struct {
    config.ConfigArgs
    Path []string `help:"Path to course JSON file." arg:"" type:"existingfile"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Add all specified courses to the system."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal("Could not load config options.", err);
    }

    db.MustOpen();
    defer db.MustClose();

    courseIDs := make([]string, 0);
    for _, path := range args.Path {
        course := db.MustAddCourse(path);
        fmt.Printf("Added course %s ('%s').\n", course.GetID(), path);
        courseIDs = append(courseIDs, course.GetID());
    }

    fmt.Printf("Added %d courses.\n", len(courseIDs));
    for _, courseID := range courseIDs {
        fmt.Printf("    %s\n", courseID);
    }
}
