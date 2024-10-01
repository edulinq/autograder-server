package main

import (
	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api/users"
	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
	cmd.CommonCMDArgs

	Table bool `help:"Output data to stdout as a TSV." default:"false"`
}

func main() {
	kong.Parse(&args,
		kong.Description("List users on the server."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	request := users.ListRequest{}

	response, err := cmd.SendCMDRequest(`users/list`, request)
	if err != nil {
		log.Fatal("Failed to send the list server users CMD request: '%w'.", err)
	}

	var responseContent users.ListResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	if args.Table {
		cmd.PrintCMDResponseFull(request, response, users.ListResponse{}, args.Verbose, func() { cmd.ListServerUsersTable(responseContent.Users) })
	} else {
		cmd.PrintCMDResponseFull(request, response, users.ListResponse{}, args.Verbose, nil)
	}
}
