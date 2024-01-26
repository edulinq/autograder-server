package main

import (
    "fmt"

    "github.com/alecthomas/kong"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/log"
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

    db.MustOpen();
    defer db.MustClose();

    fmt.Println("Basic Levels")
    log.Trace("This message is trace.");
    log.Debug("This message is debug.");
    log.Info("This message is info.");
    log.Warn("This message is warn.");
    log.Error("This message is error.");
    // log.Fatal("This message is fatal.");

    fmt.Println("Attatched Values");
    log.Info("Attatched Int.", log.NewAttr("value", 1));
    log.Info("Attatched Float64.", log.NewAttr("value", 2.3));
    log.Info("Attatched Str.", log.NewAttr("value", "Foo Bar"));
    log.Info("Attatched Bool.", log.NewAttr("value", true));
    log.Info("Attatched Err.", fmt.Errorf("This is an error!"));
}
