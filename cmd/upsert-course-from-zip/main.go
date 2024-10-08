package main

import (
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/procedures/courses"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
	Path   string `help:"Path to a zip file with one or more courses to upsert." arg:""`
	DryRun bool   `help:"Do not actually do the operation, just state what you would do." default:"false"`
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

	options := courses.CourseUpsertOptions{
		ContextUser: db.MustGetRoot(),
		DryRun:      args.DryRun,
	}

	results, err := courses.UpsertFromZipFile(args.Path, options)
	if err != nil {
		log.Fatal("Failed to add courses from zip file.", err)
	}

	fmt.Println(util.MustToJSONIndent(results))
}
