package main

import (
	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/users"
	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
	cmd.CommonOptions

	Email    string `help:"Email for the user." arg:"" required:""`
	Password string `help:"Password for the user." arg:"" required:""`
}

func main() {
	kong.Parse(&args,
		kong.Description("Authenticate as a user."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	request := users.AuthRequest{
		APIRequestUserContext: core.APIRequestUserContext{
			UserEmail: args.Email,
			UserPass:  util.Sha256HexFromString(args.Password),
		},
	}

	cmd.MustHandleCMDRequestAndExitFull(`users/auth`, request, users.AuthResponse{}, args.CommonOptions, nil)
}
