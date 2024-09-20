package main

import (
	"fmt"
	"path/filepath"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
	Out string `help:"Writes the output to the given file in JSON format."`
}

func main() {
	kong.Parse(&args,
		kong.Description("Get the autograder's version."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	if args.Out == "" {
		fmt.Printf("Short Version: %s\n", util.GetAutograderVersion())

		fullVersion := util.GetAutograderVersion() + "-" + util.GetAutograderFullVersion().Hash
		if util.GetAutograderFullVersion().Status != "" {
			fullVersion = fullVersion + "-" + util.GetAutograderFullVersion().Status
		}

		fmt.Printf("Full  Version: %s\n", fullVersion)
		fmt.Printf("API   Version: %d\n", util.MustGetAPIVersion())
		return
	}

	versionJSONPath := util.ShouldAbs(filepath.Join(util.ShouldGetThisDir(), "..", "..", args.Out))

	version := util.GetAutograderFullVersion()

	err = util.ToJSONFileIndent(&version, versionJSONPath)
	if err != nil {
		log.Error("Failed to write to the JSON file", err, log.NewAttr("path", versionJSONPath))

	}
}
