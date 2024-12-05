package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api"
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
	cmd.CommonOptions

	Endpoint   string   `help:"Endpoint of the desired API." arg:""`
	Parameters []string `help:"Parameter for the endpoint in the format 'key:value', e.g., 'id:123'." arg:"" optional:""`
	Table      bool     `help:"Attempt to output data as a TSV. Will default to JSON." default:"false"`
}

const (
	USERS   = "users"
	COURSES = "courses"
	TYPE    = "type"
)

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
		// Split the parameter into it's key and value.
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
		printFunc = printCMDResponseTable
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

	return baseDescription + endpointList.String()
}

func printCMDResponseTable(response core.APIResponse) string {
	responseContent, ok := response.Content.(map[string]any)
	if !ok {
		return ""
	}

	// Don't try to format a response that has multiple keys.
	if len(responseContent) != 1 {
		return ""
	}

	users, ok := responseContent[USERS].([]any)
	if !ok {
		return ""
	}

	firstUser, ok := users[0].(map[string]any)
	if !ok {
		return ""
	}

	var headers []string
	for key := range firstUser {
		if key == COURSES {
			continue
		}
		headers = append(headers, key)
	}

	sort.Strings(headers)

	// Add courses to the end of the slice for better readability in the output.
	_, exists := firstUser[COURSES]
	if exists {
		headers = append(headers, COURSES)
	}

	var usersTable strings.Builder
	usersTable.WriteString(strings.Join(headers, "\t"))

	lines := strings.Split(usersTable.String(), "\t")

	usersTable.WriteString("\n")

	for i, user := range users {
		userMap, ok := user.(map[string]any)
		if !ok {
			return ""
		}

		var row []string
		for _, key := range lines {
			switch value := userMap[key].(type) {
			case string:
				row = append(row, value)
			default:
				row = append(row, util.MustToJSON(value))
			}
		}

		usersTable.WriteString(strings.Join(row, "\t"))

		if i < len(users)-1 {
			usersTable.WriteString("\n")
		}
	}

	return usersTable.String()
}
