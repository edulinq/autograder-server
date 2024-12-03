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

	Endpoint   string   `help:"Endpoint of the desired API." arg:""`
	Parameters []string `help:"Parameter for the endpoint in the format 'key:value', e.g., 'id:123'." arg:"" optional:""`
	Table      bool     `help:"Output data as a TSV." default:"false"`
}

func main() {
	kong.Parse(&args,
		kong.Description(generateHelpDescription()),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Failed to load config options.", err)

		// Return to prevent further execution after log.Fatal().
		return
	}

	var endpointDescription *core.EndpointDescription

	apiDescription := api.Describe(*api.GetRoutes())
	for endpoint, requestResponse := range apiDescription.Endpoints {
		if endpoint == args.Endpoint {
			endpointDescription = &requestResponse
			break
		}
	}

	if endpointDescription == nil {
		log.Fatal("Failed to find the endpoint.", log.NewAttr("endpoint", args.Endpoint))

		// Return to prevent further execution after log.Fatal().
		return
	}

	request := map[string]any{}
	for _, arg := range args.Parameters {
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) != 2 {
			log.Fatal("Invalid parameter format: missing a colon. Expected format is 'key:value', e.g., 'id:123'.", log.NewAttr("parameter", parts))

			// Return to prevent further execution after log.Fatal().
			return
		}

		request[parts[0]] = parts[1]
	}

	var printFunc cmd.CustomResponseFormatter
	if args.Table {
		printFunc = cmd.CUSTOM_OUTPUT_MAP[args.Endpoint]
		if printFunc == nil {
			log.Fatal("Table formatting is not supported for the specified endpoint.", log.NewAttr("endpoint", args.Endpoint))

			// Return to prevent further execution after log.Fatal().
			return
		}
	}

	cmd.MustHandleCMDRequestAndExitFull(args.Endpoint, request, nil, args.CommonOptions, printFunc)
}

func generateHelpDescription() string {
	baseDescription := "Execute an API request to the specified endpoint.\n\n"

	var endpointList strings.Builder
	endpointList.WriteString("List of endpoints:\n")

	apiDescription := api.Describe(*api.GetRoutes())
	for endpoint := range apiDescription.Endpoints {
		endpointList.WriteString(fmt.Sprintf("  - %s\n", endpoint))
	}

	var customOutputEndpointList strings.Builder
	customOutputEndpointList.WriteString("Endpoints supporting TSV formatting:\n")

	for endpoint := range cmd.CUSTOM_OUTPUT_MAP {
		customOutputEndpointList.WriteString(fmt.Sprintf("  - %s\n", endpoint))
	}

	return baseDescription + endpointList.String() + customOutputEndpointList.String()
}
