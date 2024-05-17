package main

import (
    "fmt"

    "github.com/alecthomas/kong"

    "github.com/edulinq/autograder/internal/config"
    "github.com/edulinq/autograder/internal/db"
    "github.com/edulinq/autograder/internal/lms"
    "github.com/edulinq/autograder/internal/log"
    "github.com/edulinq/autograder/internal/util"
)

var args struct {
    config.ConfigArgs
    Course string `help:"ID of the course." arg:""`
    Assignment string `help:"ID of the assignment." arg:""`
}

func main() {
    kong.Parse(&args,
        kong.Description("Fetch information about a specific assignment from an LMS."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal("Could not load config options.", err);
    }

    db.MustOpen();
    defer db.MustClose();

    assignment := db.MustGetAssignment(args.Course, args.Assignment);
    course := assignment.GetCourse();

    if (assignment.GetLMSID() == "") {
        log.Fatal("Assignment has no LMS ID.", assignment);
    }

    lmsAssignment, err := lms.FetchAssignment(course, assignment.GetLMSID());
    if (err != nil) {
        log.Fatal("Could not fetch assignment.", err, assignment);
    }

    fmt.Println(util.MustToJSONIndent(lmsAssignment));
}
