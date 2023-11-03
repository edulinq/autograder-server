package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/report"
    "github.com/eriq-augustine/autograder/util"
)

var args struct {
    config.ConfigArgs
    AssignmentPath string `help:"Path to assignment JSON file." arg:"" type:"existingfile"`
    HTML bool `help:"Output report as html." default:"false"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Compile a report on the current scores in the autograder for an assignment."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    assignment := db.MustLoadAssignmentConfig(args.AssignmentPath);

    report, err := report.GetAssignmentScoringReport(assignment);
    if (err != nil) {
        log.Fatal().Err(err).Str("assignment", assignment.GetID()).Msg("Failed to get scoring report.");
    }

    if (args.HTML) {
        html, err := report.ToHTML(false);
        if (err != nil) {
            log.Fatal().Err(err).Str("assignment", assignment.GetID()).Msg("Failed to generate HTML scoring report.");
        }

        fmt.Println(html);
    } else {
        fmt.Println(util.MustToJSONIndent(report));
    }
}
