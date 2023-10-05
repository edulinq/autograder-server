package main

import (
    "fmt"
    "strings"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/canvas"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

var args struct {
    config.ConfigArgs
    AssignmentPath string `help:"Path to assignment JSON file. If specified, --course-path and --assignment-id are not required." type:"existingfile"`
    AssignmentID string `help:"The Canvas ID for an assigmnet (with --course-path, can be used instead of --assignment-path)."`
    CoursePath string `help:"Path to course JSON file (with --assignment-id, can be used instead of --assignment-path)."`
}

func main() {
    kong.Parse(&args,
        kong.Description("Fetch the grades for a specific assignment from canvas." +
            " Either --assignment-path or (--course-path and --assignment-id) are required."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    assignmentCanvasID, course, err := getAssignmentIDAndCourse(args.AssignmentPath, args.AssignmentID, args.CoursePath);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to load course/assignment information.");
    }

    grades, err := canvas.FetchAssignmentGrades(course.CanvasInstanceInfo, assignmentCanvasID);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not fetch grades.");
    }

    fmt.Println("canvas_user_id\tscore\ttime\tcomments");
    for _, grade := range grades {
        textComments := make([]string, 0, len(grade.Comments));
        for _, comment := range grade.Comments {
            textComments = append(textComments, comment.Text);
        }
        comments := strings.Join(textComments, ";");

        fmt.Printf("%s\t%s\t%s\t%s\n", grade.UserID, util.FloatToStr(grade.Score), grade.Time, comments);
    }
}

func getAssignmentIDAndCourse(assignmentPath string, assignmentID string, coursePath string) (string, *model.Course, error) {
    if (assignmentPath != "") {
        assignment := model.MustLoadAssignmentConfig(assignmentPath);
        if (assignment.CanvasID == "") {
            return "", nil, fmt.Errorf("Assignment has no Canvas ID.");
        }

        return assignment.CanvasID, assignment.Course, nil;
    }

    if ((assignmentID == "") || (coursePath == "")) {
        return "", nil, fmt.Errorf("Neither --assignment-path nor (--course-path and --assignment-id) were proveded.");
    }

    course := model.MustLoadCourseConfig(coursePath);
    if (course.CanvasInstanceInfo == nil) {
        return "", nil, fmt.Errorf("Assignment's course has no Canvas info associated with it.");
    }

    return assignmentID, course, nil;
}
