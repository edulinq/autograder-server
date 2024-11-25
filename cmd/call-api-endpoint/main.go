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
)

var args struct {
	config.ConfigArgs
	cmd.CommonOptions

	Endpoint   string   `help:"Endpoint of the desired API." arg:""`
	Parameters []string `help:"Parameter(s) for the endpoint" arg:"" optional:""`
}

func main() {
	kong.Parse(&args,
		kong.Description("Execute an API request to the specified endpoint. For more information on available API endpoints, see the API resource file at: ../resources/api.json"),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Failed to load config options.", err)
	}

	var description *core.EndpointDescription

	apiDescription := api.Describe(*api.GetRoutes())
	for endpoint, requestResponse := range apiDescription.Endpoints {
		if endpoint == args.Endpoint {
			description = &requestResponse
			break
		}
	}

	if description == nil {
		fmt.Printf("Failed to find the endpoint '%s'.", args.Endpoint)

		exit.Exit(1)
		// Return after the exit code gets set to 1 to avoid continuing on during tests.
		return
	}

	request := map[string]any{}
	for _, arg := range args.Parameters {
		// Split the parameter into it's key and value.
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) != 2 {
			log.Fatal("Invalid parameter format: missing a colon. Expected format is 'key:value'. E.g. target-email:bob@test.edulinq.org", log.NewAttr("parameter", parts))
		}

		request[parts[0]] = parts[1]
	}

	cmd.MustHandleCMDRequestAndExitFull(args.Endpoint, request, nil, args.CommonOptions, nil)
}
