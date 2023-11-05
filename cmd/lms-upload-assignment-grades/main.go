package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/lms"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

var args struct {
    config.ConfigArgs
    Course string `help:"ID of the course." arg:""`
    Assignment string `help:"ID of the assignment." arg:""`
    Grades string `help:"Path to TSV file containing 'email<TAB>score'." arg:"" type:"existingfile"`
    Force bool `help:"Ignore when there are bad users and upload all the grades for good users." short:"f" default:"false"`
    DryRun bool `help:"Do not actually upload the grades, just state what you would do." default:"false"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Upload grades for an assignment to the coure's LMS from a TSV file." +
            " Either --assignment-path or (--course-path and --assignment-id) are required."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    db.MustOpen();
    defer db.MustClose();

    assignment := db.MustGetAssignment(args.Course, args.Assignment);
    course := assignment.GetCourse();

    if (assignment.GetLMSID() == "") {
        log.Fatal().Str("assignment", assignment.FullID()).Msg("Assignment has no LMS ID.");
    }

    if (course.GetLMSAdapter() == nil) {
        log.Fatal().
            Str("course-id", course.GetID()).Str("assignment-id", assignment.GetID()).
            Msg("Course has no LMS info associated with it.");
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
        err = course.GetLMSAdapter().UpdateAssignmentScores(assignment.GetLMSID(), grades);
        if (err != nil) {
            log.Fatal().Err(err).Msg("Could not upload grades.");
        }
    }

    fmt.Printf("Uploaded %d grades.\n", len(grades));
}

func loadGrades(path string, users map[string]*usr.User, force bool) ([]*lms.SubmissionScore, error) {
    grades := make([]*lms.SubmissionScore, 0);

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

        lmsID := user.LMSID;
        if (lmsID == "") {
            message := fmt.Sprintf("User '%s' (from row (%d)) has no LMS ID.", row[0], i);

            if (force) {
                fmt.Println(message);
                continue;
            } else {
                return nil, fmt.Errorf(message);
            }
        }

        grades = append(grades, &lms.SubmissionScore{
            UserID: lmsID,
            Score: util.MustStrToFloat(row[1]),
        });
    }

    return grades, nil;
}
