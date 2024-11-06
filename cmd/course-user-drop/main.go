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

	TargetEmail string `help:"Email of the user to drop." arg:""`
	CourseID    string `help:"ID of the course." arg:""`
}

func main() {
	kong.Parse(&args,
		kong.Description("Drop a user from the course."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Failed to load config options.", err)
	}

	request := users.DropRequest{
		APIRequestCourseUserContext: core.APIRequestCourseUserContext{
			CourseID: args.CourseID,
		},
		TargetCourseUser: core.TargetCourseUser{
			Email: args.TargetEmail,
		},
	}

	cmd.MustHandleCMDRequestAndExitFull(`courses/users/drop`, request, users.DropResponse{}, args.CommonOptions, nil)
}
