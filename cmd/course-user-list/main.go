package main

import (
	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/users"
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
		kong.Description("List the users in the course."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	request := users.ListRequest{
		APIRequestCourseUserContext: core.APIRequestCourseUserContext{
			CourseID: args.CourseID,
		},
	}

	var printFunc cmd.CustomResponseFormatter = nil
	if args.Table {
		printFunc = cmd.ListCourseUsersTable
	}

	cmd.MustHandleCMDRequestAndExitFull(`courses/users/list`, request, users.ListResponse{}, args.CommonOptions, printFunc)
}
