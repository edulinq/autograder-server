package main

import (
    "fmt"

    "github.com/alecthomas/kong"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/log"
    "github.com/eriq-augustine/autograder/email"
    "github.com/eriq-augustine/autograder/report"
    "github.com/eriq-augustine/autograder/util"
)

var args struct {
    config.ConfigArgs
    Course string `help:"ID of the course." arg:""`
    Email []string `help:"Email addresses to send the report to (as HTML)." short:"e"`
    HTML bool `help:"Output report as html." default:"false"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Compile a report on the current scores in the autograder for a course."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal("Could not load config options.", err);
    }

    db.MustOpen();
    defer db.MustClose();

    course := db.MustGetCourse(args.Course);

    report, err := report.GetCourseScoringReport(course);
    if (err != nil) {
        log.Fatal("Failed to get scoring report.", log.NewAttr("course", course.GetID()), err);
    }

    if (args.HTML) {
        html, err := report.ToHTML();
        if (err != nil) {
            log.Fatal("Failed to generate HTML scoring report.", log.NewAttr("course", course.GetID()), err);
        }

        fmt.Println(html);
    } else {
        fmt.Println(util.MustToJSONIndent(report));
    }

    if (len(args.Email) > 0) {
        html, err := report.ToHTML();
        if (err != nil) {
            log.Fatal("Failed to generate HTML scoring report.", log.NewAttr("course", course.GetID()), err);
        }

        subject := fmt.Sprintf("Autograder Scoring Report for %s", course.GetName());

        err = email.Send(args.Email, subject, html, true);
        if (err != nil) {
            log.Fatal("Failed to send scoring report email.", log.NewAttr("course", course.GetID()), err);
        }
    }
}
