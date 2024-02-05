package main

import (
    "fmt"
    "strings"

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
    Email string `help:"Email of the user to fetch." arg:""`
}

func main() {
    kong.Parse(&args,
        kong.Description("Fetch the grade for a specific assignment and user from an LMS."),
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

    user, err := db.GetUser(course, args.Email);
    if (err != nil) {
        log.Fatal("Failed to fetch user.", err, assignment, log.NewUserAttr(args.Email));
    }

    if (user == nil) {
        log.Fatal("Could not find user.", err, assignment, log.NewUserAttr(args.Email));
    }

    if (user.LMSID == "") {
        log.Fatal("User does not have an LMS ID.", err, assignment, user);
    }

    score, err := lms.FetchAssignmentScore(course, assignment.GetLMSID(), user.LMSID);
    if (err != nil) {
        log.Fatal("User does not have an LMS ID.", err, assignment, user);
    }

    textComments := make([]string, 0, len(score.Comments));
    for _, comment := range score.Comments {
        textComments = append(textComments, comment.Text);
    }
    comments := strings.Join(textComments, ";");

    fmt.Println("lms_user_id\tscore\ttime\tcomments");
    fmt.Printf("%s\t%s\t%s\t%s\n", score.UserID, util.FloatToStr(score.Score), score.Time, comments);
}
