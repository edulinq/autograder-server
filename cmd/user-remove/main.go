package main

import (
	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/users"
	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
)

var args struct {
	config.ConfigArgs
	cmd.CommonOptions

	TargetEmail string `help:"Email of the user to remove." arg:""`
}

func main() {
	kong.Parse(&args,
		kong.Description("Remove a user from the server."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Failed to load config options.", err)
	}

	request := users.RemoveRequest{
		TargetUser: core.TargetServerUser{
			Email: args.TargetEmail,
		},
	}

	cmd.MustHandleCMDRequestAndExitFull(`users/remove`, request, users.RemoveResponse{}, args.CommonOptions, nil)
}
