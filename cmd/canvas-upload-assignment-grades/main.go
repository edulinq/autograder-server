package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/canvas"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

var args struct {
    config.ConfigArgs
    Grades string `help:"Path to TSV file containing 'email<TAB>score'." arg:"" type:"existingfile"`
    AssignmentPath string `help:"Path to assignment JSON file. If specified, --course-path and --assignment-id are not required." type:"existingfile"`
    AssignmentID string `help:"The Canvas ID for an assigmnet (with --course-path, can be used instead of --assignment-path)."`
    CoursePath string `help:"Path to course JSON file (with --assignment-id, can be used instead of --assignment-path)."`
    Force bool `help:"Ignore when there are bad users and upload all the grades for good users." short:"f" default:"false"`
    DryRun bool `help:"Do not actually upload the grades, just state what you would do." default:"false"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Upload grades for an assignment to canvas from a TSV file." +
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

    users, err := course.GetUsers();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to fetch autograder users.");
    }

    grades, err := loadGrades(args.Grades, users, args.Force);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not fetch grades.");
    }

    if (len(grades) == 0) {
        fmt.Println("Found no grades to upload.");
    }

    if (args.DryRun) {
        fmt.Println("Dry Run: Skipping upload.");
    } else {
        err = canvas.UpdateAssignmentGrades(course.CanvasInstanceInfo, assignmentCanvasID, grades);
        if (err != nil) {
            log.Fatal().Err(err).Msg("Could not upload grades.");
        }
    }

    fmt.Printf("Uploaded %d grades.\n", len(grades));
}

func loadGrades(path string, users map[string]*model.User, force bool) ([]*canvas.CanvasGradeInfo, error) {
    grades := make([]*canvas.CanvasGradeInfo, 0);

    rows, err := util.ReadSeparatedFile(path, "\t", 0);
    if (err != nil) {
        return nil, err;
    }

    for i, row := range rows {
        if (len(row) < 2) {
            return nil, fmt.Errorf("Row (%d) does not have enough values. Expecting 2, found %d.", i, len(row));
        }

        user := users[row[0]];
        if (user == nil) {
            message := fmt.Sprintf("Row (%d) has an unrecognized user: '%s'.", i, row[0]);

            if (force) {
                fmt.Println(message);
                continue;
            } else {
                return nil, fmt.Errorf(message);
            }
        }

        canvasID := user.CanvasID;
        if (canvasID == "") {
            message := fmt.Sprintf("User '%s' (from row (%d)) has no Canvas ID.", row[0], i);

            if (force) {
                fmt.Println(message);
                continue;
            } else {
                return nil, fmt.Errorf(message);
            }
        }

        grades = append(grades, &canvas.CanvasGradeInfo{
            UserID: canvasID,
            Score: util.MustStrToFloat(row[1]),
        });
    }

    return grades, nil;
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
