package main

import (
	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/assignments"
	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
)

var args struct {
	config.ConfigArgs
	cmd.CommonOptions

	CourseID string `help:"ID of the course." arg:""`
	Table    bool   `help:"Output data as a TSV." default:"false"`
}

func main() {
	kong.Parse(&args,
		kong.Description("List the assignments from the course."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	request := assignments.ListRequest{
		APIRequestCourseUserContext: core.APIRequestCourseUserContext{
			CourseID: args.CourseID,
		},
	}

	cmd.MustHandleCMDRequestAndExitFull(`courses/assignments/list`, request, assignments.ListResponse{}, args.CommonOptions, nil)
}
