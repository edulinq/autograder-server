package main

import (
    "fmt"
    "os"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/canvas"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
)

var args struct {
    config.ConfigArgs
    Path string `help:"Path to course JSON file." arg:"" type:"existingfile"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Sync IDs with matching canvas users (does not add/remove users)."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    course := model.MustLoadCourseConfig(args.Path);
    if (course.CanvasInstanceInfo == nil) {
        fmt.Println("Course has no Canvas info associated with it.");
        os.Exit(2);
    }

    users, err := course.GetUsers();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to fetch autograder users.");
    }

    canvasUsers, err := canvas.FetchUsers(course.CanvasInstanceInfo);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to fetch canvas users.");
    }

    count := 0
    for _, canvasUser := range canvasUsers {
        user := users[canvasUser.Email]
        if (user == nil) {
            continue;
        }

        if (user.CanvasID == canvasUser.ID) {
            continue;
        }

        user.CanvasID = canvasUser.ID;
        count++;
    }

    err = course.SaveUsersFile(users);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to save users file.");
    }

    fmt.Printf("Updated %d users.\n", count);
}
