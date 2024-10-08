package main

import (
	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api/users"
	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
)

var args struct {
	config.ConfigArgs
	cmd.CommonOptions

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

	var printFunc cmd.CMDPrintHandler = nil
	if args.Table {
		printFunc = cmd.ListServerUsersTable
	}

	cmd.MustHandleCMDRequestAndExitFull(`users/list`, request, users.ListResponse{}, args.CommonOptions, printFunc)
}
