package main

import (
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
	Source string `help:"The source to add a course from." arg:""`
}

func main() {
	kong.Parse(&args,
		kong.Description("Add a course to system from a source (FileSpec)."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	db.MustOpen()
	defer db.MustClose()

	spec, err := common.ParseFileSpec(args.Source)
	if err != nil {
		log.Fatal("Failed to parse FileSpec.", err)
	}

	tempDir, err := util.MkDirTemp("autograder-add-course-source-")
	if err != nil {
		log.Fatal("Failed to make temp source dir.", err)
	}
	defer util.RemoveDirent(tempDir)

	err = spec.CopyTarget(common.ShouldGetCWD(), tempDir, false)
	if err != nil {
		log.Fatal("Failed to copy source.", err)
	}

	courseIDs, err := db.AddCoursesFromDir(tempDir, spec)
	if err != nil {
		log.Fatal("Failed to add course dir.", err)
	}

	fmt.Printf("Added %d courses.\n", len(courseIDs))
	for _, courseID := range courseIDs {
		fmt.Printf("    %s\n", courseID)
	}
}
