package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

var args struct {
    config.ConfigArgs
    AssignmentPath string `help:"Path to assignment JSON file." arg:"" type:"existingfile"`
    DryRun bool `help:"Do not actually upload the grades, just state what you would do." default:"false"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Upload grades after late polices have been applied for an assignment to canvas from local submissions."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    assignment := model.MustLoadAssignmentConfig(args.AssignmentPath);
    if (assignment.CanvasID == "") {
        log.Fatal().Msg("Assignment has no Canvas ID.");
    }

    if (assignment.Course.CanvasInstanceInfo == nil) {
        log.Fatal().Msg("Assignment's course has no Canvas info associated with it.");
    }

    users, err := assignment.Course.GetUsers();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to fetch autograder users.");
    }

    scoringInfos, err := assignment.GetScoringInfo(users);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to get scoring information.");
    }

    err = assignment.LatePolicy.Apply(assignment, scoringInfos, args.DryRun);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to apply late policy.");
    }

    if (args.DryRun) {
        fmt.Println("Dry Run: Showing scoring infos instead of uploading them.");
        fmt.Println(util.MustToJSONIndent(scoringInfos));
    }

    // TEST
    fmt.Println("Uploaded assignment grades.");
}
