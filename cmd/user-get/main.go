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

	TargetEmail string `help:"Email of the user to get." arg:""`
}

func main() {
	kong.Parse(&args,
		kong.Description("Get the information for a server user."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	request := users.GetRequest{
		TargetUser: core.TargetServerUserSelfOrAdmin{
			TargetServerUser: core.TargetServerUser{
				Email: args.TargetEmail,
			},
		},
	}

	cmd.MustHandleCMDRequestAndExitFull(`users/get`, request, users.GetResponse{}, args.CommonOptions, nil)
}
