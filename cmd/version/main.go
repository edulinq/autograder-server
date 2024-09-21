package main

import (
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
	Out string `help:"If provided, the output will be written to the specified file in JSON format."`
}

func main() {
	kong.Parse(&args,
		kong.Description("Get the autograder's version."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	version, err := util.GetAutograderVersion()
	if err != nil {
		log.Error("Failed to get the autograder version.", err)
	}

	if args.Out == "" {
		fmt.Printf("Short Version: %s\n", version.Short)
		fmt.Printf("Full  Version: %s\n", version.String())
		fmt.Printf("API   Version: %d\n", version.Api)

		return
	}

	err = util.ToJSONFileIndent(version, args.Out)
	if err != nil {
		log.Fatal("Failed to write to the JSON file", err, log.NewAttr("path", args.Out))
	}
}
