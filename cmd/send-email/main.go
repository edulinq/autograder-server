package main

import (
    "github.com/alecthomas/kong"

    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/email"
    "github.com/edulinq/autograder/log"
    "github.com/edulinq/autograder/db"
)

var args struct {
    config.ConfigArgs
    To []string `help:"Email recipents." required:""`
    Subject string `help:"Email subject." required:""`
    Body string `help:"Email body." required:""`
    //TODO: fix this to be an optional argument
    Course string `help:"Course ID." default:""`
}

func main() {
    kong.Parse(&args,
        kong.Description("Send an email."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal("Could not load config options.", err);
    }

    // If database cannot get course and using role based addressing
    // Then, resolve users will return an error
    // So, we do not error check on the get course method
    course, err := db.GetCourse(args.Course);
    emailTo, err := db.ResolveUsers(course, args.To);
    if (err != nil) {
        log.Fatal("Failed to resolve users:'%s', '%w'.", course.GetName(), err);
    }
    if (err != nil) {
        log.Fatal("Could not resolve users.", err);
    }
    err = email.Send(emailTo, args.Subject, args.Body, false);
    if (err != nil) {
        log.Fatal("Could not send email.", err);
    }
}
