package main

import (
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api/users"
	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs

	Table   bool `help:"Output data to stdout as a TSV." default:"false"`
	Verbose bool `help:"Use verbose output to show full request/response without specific formatting." default:"false"`
}

func main() {
	kong.Parse(&args,
		kong.Description("List users on the server."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	err = listServerUsers(args.Table)
	if err != nil {
		log.Fatal("Failed to list server users.", err)
	}
}

func listServerUsers(table bool) error {
	request := users.ListRequest{}

	response, err := cmd.SendCMDRequest(`users/list`, request)
	if err != nil {
		return fmt.Errorf("Failed to send the list server users CMD request: '%w'.", err)
	}

	var responseContent users.ListResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	if config.TESTING_MODE.Get() {
		cmd.PrintCMDResponse(request, response, users.ListResponse{}, args.Verbose)
		return nil
	}

	if args.Verbose {
		fmt.Printf("\nAutograder Request:\n---\n%s\n---\n", util.MustToJSONIndent(request))
		fmt.Printf("\nAutograder Response:\n---\n%s\n---\n", util.MustToJSONIndent(response))
	}

	cmd.ListUsers(responseContent.Users, args.Table)

	return nil
}
