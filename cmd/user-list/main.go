package main

import (
	"strings"

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

	Table bool `help:"Output data as a TSV." default:"false"`
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

	var printFunc cmd.CustomResponseFormatter = nil
	if args.Table {
		printFunc = listServerUsersTable
	}

	cmd.MustHandleCMDRequestAndExitFull(`users/list`, request, users.ListResponse{}, args.CommonOptions, printFunc)
}

func listServerUsersTable(response core.APIResponse) string {
	var responseContent users.ListResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	var serverUsersTable strings.Builder

	headers := []string{"email", "name", "server-role", "courses"}
	serverUsersTable.WriteString(strings.Join(headers, "\t") + "\n")

	for i, user := range responseContent.Users {
		if i > 0 {
			serverUsersTable.WriteString("\n")
		}

		serverUsersTable.WriteString(user.Email)
		serverUsersTable.WriteString("\t")
		serverUsersTable.WriteString(user.Name)
		serverUsersTable.WriteString("\t")
		serverUsersTable.WriteString(user.Role.String())
		serverUsersTable.WriteString("\t")
		serverUsersTable.WriteString(util.MustToJSONIndent(user.Courses))
	}

	return serverUsersTable.String()
}
