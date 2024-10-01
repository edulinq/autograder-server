package main

import (
	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/assignments/submissions"
	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
)

var args struct {
	config.ConfigArgs
	cmd.CommonCMDArgs

	TargetEmail      string `help:"Email of the user to fetch." arg:""`
	CourseID         string `help:"ID of the course." arg:""`
	AssignmentID     string `help:"ID of the assignment." arg:""`
	TargetSubmission string `help:"ID of the submission. Defaults to the latest submission." arg:"" optional:""`
}

func main() {
	kong.Parse(&args,
		kong.Description("Fetch a submission for a specific assignment and user."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Failed to load config options.", err)
	}

	request := submissions.FetchUserPeekRequest{
		APIRequestAssignmentContext: core.APIRequestAssignmentContext{
			APIRequestCourseUserContext: core.APIRequestCourseUserContext{
				CourseID: args.CourseID,
			},
			AssignmentID: args.AssignmentID,
		},
		TargetUser: core.TargetCourseUserSelfOrGrader{
			TargetCourseUser: core.TargetCourseUser{
				Email: args.TargetEmail,
			},
		},
		TargetSubmission: args.TargetSubmission,
	}

	err = cmd.SendAndPrintCMDRequest(`courses/assignments/submissions/fetch/user/peek`, request, submissions.FetchUserPeekResponse{}, args.Verbose, nil)
	if err != nil {
		log.Fatal("Failed to peek the user's submission.", err)
	}
}
