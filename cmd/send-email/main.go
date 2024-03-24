package main

import (
    "github.com/alecthomas/kong"

    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/db"
    "github.com/edulinq/autograder/email"
    "github.com/edulinq/autograder/log"
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

    if (args.Course != "") {
        course := db.MustGetCourse(args.Course);

        args.To, err = db.ResolveUsers(course, args.To);
        if (err != nil) {
            log.Fatal("Failed to resolve users: '%w'.", err);
        }
    }

    err = email.Send(args.To, args.Subject, args.Body, false);
    if (err != nil) {
        log.Fatal("Could not send email.", err);
    }
}
