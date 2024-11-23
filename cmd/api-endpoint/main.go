package main

import (
	"fmt"

	"strings"

	"github.com/alecthomas/kong"
	"github.com/edulinq/autograder/internal/api"
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/exit"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
	cmd.CommonOptions

	Endpoint   string   `help:"Endpoint of the desired action." arg:""`
	Parameters []string `help:"Parameters for the endpoint" arg:"" optional:""`
	Table      bool     `help:"Output data as a TSV." default:"false"`
}

func main() {
	kong.Parse(&args,
		kong.Description("Perform an action with the desired endpoint."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Failed to load config options.", err)
	}

	var description core.EndpointDescription

	describe := api.Describe(*api.GetRoutes())
	for endpoint, requestResponse := range describe.Endpoints {
		if endpoint == args.Endpoint {
			description = requestResponse
			break
		}

	}

	if description == (core.EndpointDescription{}) {
		log.Error("Failed to find the endpoint.", log.NewAttr("endpoint", args.Endpoint))

		if !config.UNIT_TESTING_MODE.Get() {
			fmt.Print(util.MustToJSONIndent(api.Describe(*api.GetRoutes())))
		}

		exit.Exit(1)
	}

	var printFunc cmd.CustomResponseFormatter
	if args.Table {
		switch args.Endpoint {
		case usersList:
			printFunc = cmd.ListServerUsersTable
		case courseUserList:
			printFunc = cmd.ListCourseUsersTable
		default:
			printFunc = nil
		}
	}

	request := map[string]any{}

	for _, arg := range args.Parameters {
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) < 2 {
			log.Fatal("No colon provided.")
		}

		request[parts[0]] = parts[1]
	}

	cmd.MustHandleCMDRequestAndExitFull(args.Endpoint, request, "", args.CommonOptions, printFunc)
}

const usersList = "users/list"
const courseUserList = "courses/users/list"
