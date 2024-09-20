package main

import (
	"fmt"
	"path/filepath"
	"strings"

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
		fmt.Printf("Full  Version: %s\n", util.GetAutograderFullVersion())
		fmt.Printf("API   Version: %d\n", util.MustGetAPIVersion())
		return
	}

	versionJSONPath := util.ShouldAbs(filepath.Join(util.ShouldGetThisDir(), "..", "..", args.Out))

	versionSplice := strings.Split(util.GetAutograderFullVersion(), "-")

	version := util.Version{
		Short:  util.GetAutograderVersion(),
		Hash:   versionSplice[1],
		Status: versionSplice[2],
		Api:    util.MustGetAPIVersion(),
	}

	err = util.ToJSONFileIndent(&version, versionJSONPath)
	if err != nil {
		log.Error("Failed to write to the JSON file", err, log.NewAttr("path", versionJSONPath))

	}
}
