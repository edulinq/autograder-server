package main

import (
    "github.com/alecthomas/kong"

    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/db"
    "github.com/edulinq/autograder/email"
    "github.com/edulinq/autograder/log"
    "github.com/edulinq/autograder/model"
)

var args struct {
    config.ConfigArgs
    Course string `help:"Optional Course ID. The course is only required when emailing by role or the *, which denotes all roles." arg:"" optional:""`
    To []string `help:"Email recipents." required:""`
    Subject string `help:"Email subject." required:""`
    Body string `help:"Email body." required:""`
}

func main() {
    kong.Parse(&args,
        kong.Description("Send an email."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal("Could not load config options.", err);
    }

    var course *model.Course;
    if (args.Course != "") {
        course, err = db.GetCourse(args.Course);
        if (err != nil) {
            log.Fatal("Failed to get course: '%s', '%w'.", args.Course, err);
        }
    } else {
        course = nil;
    }

    emailTo, err := db.ResolveUsers(course, args.To);
    if (err != nil) {
        log.Fatal("Failed to resolve users: '%w'.", err);
    }

    err = email.Send(emailTo, args.Subject, args.Body, false);
    if (err != nil) {
        log.Fatal("Could not send email.", err);
    }
}
