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
	Course string `help:"ID of the course." arg:""`
	DryRun bool   `help:"Do not actually do the operation, just state what you would do." default:"false"`
}

func main() {
	kong.Parse(&args,
		kong.Description("Update a course with the existing (or new) source."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	db.MustOpen()
	defer db.MustClose()

	course := db.MustGetCourse(args.Course)

	options := courses.CourseUpsertOptions{
		ContextUser: db.MustGetRoot(),
		DryRun:      args.DryRun,
	}

	result, err := courses.UpdateFromLocalSource(course, options)
	if err != nil {
		log.Fatal("Failed to update course.", err, course)
	}

	fmt.Println(util.MustToJSONIndent(result))
}
