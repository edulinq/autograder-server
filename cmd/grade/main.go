package main

import (
    "fmt"

    "github.com/alecthomas/kong"

    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/db"
    "github.com/edulinq/autograder/log"
    "github.com/edulinq/autograder/grader"
    "github.com/edulinq/autograder/util"
)

var args struct {
    config.ConfigArgs
    Course string `help:"ID of the course." arg:""`
    Assignment string `help:"ID of the assignment." arg:""`
    Submission string `help:"Path to submission directory." required:"" type:"existingdir"`
    OutPath string `help:"Option path to output a JSON grading result." type:"path"`
    User string `help:"User email for the submission." default:"testuser"`
    Message string `help:"Submission message." default:""`
}

func main() {
    kong.Parse(&args);
    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal("Could not load config options.", err);
    }

    db.MustOpen();
    defer db.MustClose();

    assignment := db.MustGetAssignment(args.Course, args.Assignment);

    result, reject, err := grader.GradeDefault(assignment, args.Submission, args.User, args.Message);
    if (err != nil) {
        if (result.HasTextOutput()) {
            fmt.Println("Grading failed, but output was recovered:");
            fmt.Println(result.GetCombinedOutput());
        }
        log.Fatal("Failed to run grader.", assignment, err);
    }

    if (reject != nil) {
        log.Fatal("Submission was rejected.", assignment, log.NewAttr("reject-reason", reject.String()));
    }

    if (args.OutPath != "") {
        err = util.ToJSONFileIndent(result.Info, args.OutPath);
        if (err != nil) {
            log.Fatal("Failed to output JSON result.", assignment, log.NewAttr("outpath", args.OutPath), err);
        }
    }

    fmt.Println(result.Info.Report());
}
