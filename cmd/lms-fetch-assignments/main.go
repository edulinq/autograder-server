package main

import (
    "fmt"

    "github.com/alecthomas/kong"

    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/db"
    "github.com/edulinq/autograder/lms"
    "github.com/edulinq/autograder/log"
    "github.com/edulinq/autograder/util"
)

var args struct {
    config.ConfigArgs
    Course string `help:"ID of the course." arg:""`
}

func main() {
    kong.Parse(&args,
        kong.Description("Fetch information about all assignments from an LMS."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal("Could not load config options.", err);
    }

    db.MustOpen();
    defer db.MustClose();

    course := db.MustGetCourse(args.Course);

    lmsAssignments, err := lms.FetchAssignments(course);
    if (err != nil) {
        log.Fatal("Failed to fetch assignments.", err, course);
    }

    fmt.Println(util.MustToJSONIndent(lmsAssignments));
}
