package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api"
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/exit"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
	cmd.CommonOptions

	EndpointSubstring string   `help:"Substring of the desired API endpoint. Must match exactly one endpoint. If this matches the full text of an endpoint, then substring matches will be ignored." arg:"" optional:""`
	Parameters        []string `help:"Parameter for the endpoint in the format 'key:value', e.g., 'id:123'." arg:"" optional:""`
	ExactMatch        bool     `help:"Don't check for substring endpoint matches." default:"false"`
	Table             bool     `help:"Attempt to output data as a TSV. Fallback to JSON if the table conversion fails." default:"false"`
	List              bool     `help:"List all API endpoints." default:"false" short:"l"`
	Describe          bool     `help:"Describe the given endpoint instead of calling it (e.g., output information about the parameters)." default:"false" short:"d"`
}

func main() {
	kong.Parse(&args,
		kong.Description("Call an endpoint."),
	)

	// Note the return/exit behavior in this function.
	// During testing log.Fatal() and exit.Exit() will not actually terminate the program,
	// but they will set an exit status that will be retrieved and compared in testing.
	// So all exiting cases will need to make sure to return, even if after a log.Fatal() or exit.Exit().

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Failed to load config options.", err)
		return
	}

	if args.List {
		listAPIEndpoints()
		return
	}

	if args.EndpointSubstring == "" {
		log.Error("Please enter an endpoint. Use --list to view all endpoints.")
		exit.Exit(1)
		return
	}

	apiDescription, err := core.DescribeRoutes(*api.GetRoutes())
	if err != nil {
		log.Fatal("Failed to describe API endpoints.", err)
		return
	}

	endpoint := matchEndpoint(apiDescription)
	if endpoint == "" {
		return
	}

	if args.Describe {
		fmt.Println(util.MustToJSONIndent(apiDescription.Endpoints[endpoint]))
		return
	}

	request := make(map[string]any, 0)

	for _, arg := range args.Parameters {
		// Split the parameter into it's key and value.
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) != 2 {
			log.Fatal("Invalid parameter format: missing a colon. Expected format is 'key:value', e.g., 'id:123'.", log.NewAttr("parameter", parts))
			return
		}

		request[parts[0]] = parts[1]
	}

	err = updateParams(apiDescription.Endpoints[endpoint], request)
	if err != nil {
		log.Fatal("Failed to parse endpoint-specific params.", err)
		return
	}

	// Add in core request fields.
	request["source"] = common.AG_REQUEST_SOURCE
	request["source-version"] = util.MustGetFullCachedVersion().String()

	var printFunc cmd.CustomResponseFormatter = nil
	if args.Table {
		printFunc = cmd.ConvertAPIResponseToTable
	}

	cmd.MustHandleCMDRequestAndExitFull(args.EndpointSubstring, request, nil, args.CommonOptions, printFunc)
}

// Update the parameters that will be sent to the server according to the specific endpoint.
// This is where parameters can be typed.
func updateParams(endpoint core.EndpointDescription, params map[string]any) error {
	var allErrors error

	for _, field := range endpoint.Input {
		raw_value, exists := params[field.Name]
		if !exists {
			continue
		}

		// If the value is not a string, then it must gave already been handled.
		string_value, ok := raw_value.(string)
		if !ok {
			continue
		}

		if field.Type == "bool" {
			string_value = strings.TrimSpace(string_value)
			if string_value == "true" {
				params[field.Name] = true
			} else if string_value == "false" {
				params[field.Name] = false
			} else {
				allErrors = errors.Join(allErrors, fmt.Errorf("Param '%s': Failed to convert boolean: '%v'.", field.Name, raw_value))
				continue
			}
		}
	}

	return allErrors
}

func listAPIEndpoints() {
	apiDescription, err := core.DescribeRoutes(*api.GetRoutes())
	if err != nil {
		log.Fatal("Failed to describe API endpoints: '%w'.", err)
	}

	var endpoints []string
	for endpoint, _ := range apiDescription.Endpoints {
		endpoints = append(endpoints, endpoint)
	}

	slices.Sort(endpoints)

	for _, endpoint := range endpoints {
		fmt.Println(endpoint)
	}
}

func matchEndpoint(apiDescription *core.APIDescription) string {
	substringMatches := make([]string, 0)
	exactMatch := ""

	for endpoint, _ := range apiDescription.Endpoints {
		if endpoint == args.EndpointSubstring {
			exactMatch = endpoint
		}

		if strings.Contains(endpoint, args.EndpointSubstring) {
			substringMatches = append(substringMatches, endpoint)
		}
	}

	if exactMatch != "" {
		return exactMatch
	}

	if args.ExactMatch {
		log.Fatal("Failed to find an exact endpoint match. Use --list to view all endpoints.",
			log.NewAttr("endpoint-substring", args.EndpointSubstring))
		return ""
	}

	if len(substringMatches) == 0 {
		log.Fatal("Failed to find matching endpoint. Use --list to view all endpoints.",
			log.NewAttr("endpoint-substring", args.EndpointSubstring))
		return ""
	}

	if len(substringMatches) > 1 {
		slices.Sort(substringMatches)
		log.Fatal("Found multiple matching endpoints. Use --list to view all endpoints.",
			log.NewAttr("endpoint-substring", args.EndpointSubstring),
			log.NewAttr("matching-endpoints", substringMatches))
		return ""
	}

	return substringMatches[0]
}
