package main

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api"
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
)

var args struct {
	config.ConfigArgs
	cmd.CommonOptions

	Endpoint   string   `help:"Endpoint of the desired API." args:""`
	Parameters []string `help:"Parameter for the endpoint in the format 'key:value', e.g., 'id:123'." arg:"" optional:""`
	Table      bool     `help:"Attempt to output data as a TSV. Fallback to JSON if the table conversion fails." default:"false"`
	List       bool     `help:"List all endpoints." default:"false"`
}

func main() {
	kong.Parse(&args,
		kong.Description("Execute an API request to the specified endpoint."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Failed to load config options.", err)

		// Return to prevent further execution after log.Fatal().
		return
	}

	if args.List {
		apiDescription := api.Describe(*api.GetRoutes())
		for endpoint := range apiDescription.Endpoints {
			fmt.Println(endpoint)
		}

		return
	}


	var endpointDescription *core.EndpointDescription

	apiDescription, err := api.Describe(*api.GetRoutes())
	if err != nil {
		log.Fatal("Failed to describe API endpoints.", err)
	}

	for endpoint, requestResponse := range apiDescription.Endpoints {
		if endpoint == args.Endpoint {
			endpointDescription = &requestResponse
			break
		}
	}

	if endpointDescription == nil {
		log.Fatal("Failed to find the endpoint. See --list to view all endpoints.", log.NewAttr("endpoint", args.Endpoint))

		// Return to prevent further execution after log.Fatal().
		return
	}

	request := map[string]any{}
	for _, arg := range args.Parameters {
		// Split the parameter into it's key and value.
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) != 2 {
			log.Fatal("Invalid parameter format: missing a colon. Expected format is 'key:value', e.g., 'id:123'.", log.NewAttr("parameter", parts))

			// Return to prevent further execution after log.Fatal().
			return
		}

		request[parts[0]] = parts[1]
	}

	var printFunc cmd.CustomResponseFormatter = nil
	if args.Table {
		printFunc = cmd.ConvertAPIResponseToTable
	}

	cmd.MustHandleCMDRequestAndExitFull(args.Endpoint, request, nil, args.CommonOptions, printFunc)
}
