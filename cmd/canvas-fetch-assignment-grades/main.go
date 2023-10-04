package main

import (
    "fmt"
    "os"
    "strings"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/canvas"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
)

var args struct {
    config.ConfigArgs
    Path string `help:"Path to assignment JSON file." arg:"" type:"existingfile"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Fetch the grades for a specific assignment from canvas."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    assignment := model.MustLoadAssignmentConfig(args.Path);
    if (assignment.Course.CanvasInfo == nil) {
        fmt.Println("Assignment's course has no Canvas info associated with it.");
        os.Exit(2);
    }

    if (assignment.CanvasID == "") {
        fmt.Println("Assignment has no Canvas ID.");
        os.Exit(3);
    }

    grades, err := canvas.FetchAssignmentGrades(assignment.Course.CanvasInfo, assignment.CanvasID);
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

        fmt.Printf("%s\t%s\t%s\t%s\n", grade.UserID, grade.Score, grade.Time, comments);
    }
}
