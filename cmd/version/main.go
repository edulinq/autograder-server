package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
	OUT string `help:"Redirects the output to the given file in JSON format." default:""`
}

type Version struct {
	Short  string `json:"short-version"`
	Hash   string `json:"git-hash"`
	Status string `json:"status"`
}

func main() {
	kong.Parse(&args,
		kong.Description("Get the autograder's version."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	if !(args.OUT == "") {
		versionJSONPath := util.ShouldAbs(filepath.Join(util.ShouldGetThisDir(), "..", "..", args.OUT))

		versionSplice := strings.Split(util.GetAutograderFullVersion(), "-")

		version := Version{
			Short:  util.GetAutograderVersion(),
			Hash:   versionSplice[1],
			Status: versionSplice[2],
		}

		err := util.ToJSONFileIndentCustom(&version, versionJSONPath, "", " ")
		if err != nil {
			log.Error("Failed to write to the JSON file", err, log.NewAttr("path", versionJSONPath))

		}

		return
	}

	fmt.Println("Autograder")
	fmt.Printf("Short Version: %s\n", util.GetAutograderVersion())
	fmt.Printf("Full  Version: %s\n", util.GetAutograderFullVersion())
	fmt.Printf("API   Version: %d\n", core.API_VERSION)
}
