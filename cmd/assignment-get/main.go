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

	CourseID     string `help:"ID of the course." arg:""`
	AssignmentID string `help:"ID of the assignment." arg:""`
}

func main() {
	kong.Parse(&args,
		kong.Description("Get the information for a course assignment."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Failed to load config options.", err)
	}

	request := assignments.GetRequest{
		APIRequestAssignmentContext: core.APIRequestAssignmentContext{
			APIRequestCourseUserContext: core.APIRequestCourseUserContext{
				CourseID: args.CourseID,
			},
			AssignmentID: args.AssignmentID,
		},
	}

	cmd.MustHandleCMDRequestAndExitFull(`courses/assignments/get`, request, assignments.GetResponse{}, args.CommonOptions, nil)
}
