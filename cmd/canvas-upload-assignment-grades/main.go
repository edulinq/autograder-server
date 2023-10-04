package main

import (
    "fmt"
    "os"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/canvas"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

var args struct {
    config.ConfigArgs
    Assignment string `help:"Path to assignment JSON file." arg:"" type:"existingfile"`
    Grades string `help:"Path to TSV file containing 'email<TAB>score'." arg:"" type:"existingfile"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Upload grades for an assignment to canvas from a TSV file."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    assignment := model.MustLoadAssignmentConfig(args.Assignment);
    if (assignment.Course.CanvasInfo == nil) {
        fmt.Println("Assignment's course has no Canvas info associated with it.");
        os.Exit(2);
    }

    if (assignment.CanvasID == "") {
        fmt.Println("Assignment has no Canvas ID.");
        os.Exit(3);
    }

    users, err := assignment.Course.GetUsers();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to fetch autograder users.");
    }

    grades, err := loadGrades(args.Grades, users);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not fetch grades.");
    }

    if (len(grades) == 0) {
        fmt.Println("Found no grades to upload.");
    }

    err = canvas.UpdateAssignmentGrades(assignment.Course.CanvasInfo, assignment.CanvasID, grades);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not upload grades.");
    }

    fmt.Println("Grades uploaded.");
}

func loadGrades(path string, users map[string]*model.User) ([]model.CanvasGradeInfo, error) { 
    grades := make([]model.CanvasGradeInfo, 0);

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
            return nil, fmt.Errorf("Row (%d) has an unrecognized user: '%s'.", i, row[0]);
        }

        canvasID := user.CanvasID;
        if (canvasID == "") {
            return nil, fmt.Errorf("User '%s' (from row (%d)) has no Canvas ID.", row[0], i);
        }

        grades = append(grades, model.CanvasGradeInfo{
            UserID: canvasID,
            Score: row[1],
        });
    }

    return grades, nil;
}
