package main

import (
    "fmt"
    "os"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
)

var args struct {
    config.ConfigArgs
    Path string `help:"Path to course JSON file." arg:"" type:"existingfile"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Fetch users for a specific LMS course."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    course := db.MustLoadCourseConfig(args.Path);
    if (course.GetLMSAdapter() == nil) {
        fmt.Println("Course has no LMS info associated with it.");
        os.Exit(2);
    }

    users, err := course.GetLMSAdapter().FetchUsers();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not fetch users.");
    }

    fmt.Println("id\temail\tname\trole");
    for _, user := range users {
        fmt.Printf("%s\t%s\t%s\t%s\n", user.ID, user.Email, user.Name, user.Role.String());
    }
}
