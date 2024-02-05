package main

import (
    "fmt"

    "github.com/alecthomas/kong"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/lms"
    "github.com/eriq-augustine/autograder/log"
    "github.com/eriq-augustine/autograder/util"
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
