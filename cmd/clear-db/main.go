package main

import (
    "github.com/alecthomas/kong"

    "github.com/edulinq/autograder/internal/config"
    "github.com/edulinq/autograder/internal/db"
    "github.com/edulinq/autograder/internal/log"
)

var args struct {
    config.ConfigArgs
}

func main() {
    kong.Parse(&args,
        kong.Description("Clear the current database. Be careful."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal("Could not load config options.", err);
    }

    db.MustOpen();
    defer db.MustClose();

    err = db.Clear();
    if (err != nil) {
        log.Fatal("Failed to clear database.", err);
    }

}
