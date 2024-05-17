package main

import (
    "fmt"

    "github.com/alecthomas/kong"

    "github.com/edulinq/autograder/internal/config"
    "github.com/edulinq/autograder/internal/log"
)

var args struct {
    config.ConfigArgs
}

func main() {
    kong.Parse(&args,
        kong.Description("Dump all the loaded config and exit."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal("Could not load config options.", err);
    }

    jsonText, err := config.ToJSON();
    if (err != nil) {
        log.Fatal("Could not serialize config.", err);
    }

    fmt.Println(jsonText);
}
