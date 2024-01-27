package main

import (
    "fmt"
    "time"

    "github.com/alecthomas/kong"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/log"
    "github.com/eriq-augustine/autograder/util"
)

var args struct {
    config.ConfigArgs

    Level string `help:"Only includes logs from this level or higher." short:"l" default:"info"`
    Time string `help:"Only includes logs from this time or later." short:"t"`

    Course string `help:"Only includes logs from this course."`
    Assignment string `help:"Only includes logs from this assignment." short:"a"`
    User string `help:"Only includes logs from this user." short:"u"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Dump all the loaded config and exit."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal("Could not load config options.", err);
    }

    db.MustOpen();
    defer db.MustClose();

    level, err := log.ParseLevel(args.Level);
    if (err != nil) {
        log.Fatal("Could not parse logging level.", err);
    }

    after := time.Time{};
    if (args.Time != "") {
        after, err = util.GuessTime(args.Time);
        if (err != nil) {
            log.Fatal("Could not parse time.", err);
        }
    }

    logs, err := db.GetLogRecords(level, after, args.Course, args.Assignment, args.User);
    if (err != nil) {
        log.Fatal("Failed to fetch logs.", err);
    }

    for _, log := range logs {
        fmt.Println(log.String());
    }
}
