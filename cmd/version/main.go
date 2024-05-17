package main

import (
    "fmt"

    "github.com/alecthomas/kong"

    "github.com/edulinq/autograder/internal/api/core"
    "github.com/edulinq/autograder/internal/config"
    "github.com/edulinq/autograder/internal/log"
    "github.com/edulinq/autograder/internal/util"
)

var args struct {
    config.ConfigArgs
}

func main() {
    kong.Parse(&args,
        kong.Description("Get the autograder's version."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal("Could not load config options.", err);
    }

    fmt.Println("Autograder");
    fmt.Printf("Short Version: %s\n", util.GetAutograderVersion());
    fmt.Printf("Full  Version: %s\n", util.GetAutograderFullVersion());
    fmt.Printf("API   Version: %d\n", core.API_VERSION);
}
