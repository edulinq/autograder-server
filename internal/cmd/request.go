package cmd

import (
	"fmt"
	"net"
	"reflect"
	"sort"
	"strings"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/server"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/exit"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

type CommonOptions struct {
	Verbose bool `help:"Add the full request and response to the output. Be aware that the output will include extra text beyond the expected format." default:"false"`
}

type CustomResponseFormatter func(response core.APIResponse) (string, bool)

func MustHandleCMDRequestAndExit(endpoint string, request any, responseType any) {
	MustHandleCMDRequestAndExitFull(endpoint, request, responseType, CommonOptions{}, nil)
}

func MustHandleCMDRequestAndExitFull(endpoint string, request any, responseType any, options CommonOptions, customPrintFunc CustomResponseFormatter) {
	var response core.APIResponse
	var err error

	// Run inside a func so defers will run before exit.Exit().
	func() {
		startedCMDServer, oldPort := mustEnsureServerIsRunning()
		if startedCMDServer {
			defer server.StopServer(true)
			defer config.WEB_PORT.Set(oldPort)
		}

		response, err = SendCMDRequest(endpoint, request)
	}()

	if err != nil {
		log.Fatal("Failed to send the CMD request.", err, log.NewAttr("endpoint", endpoint))
	}

	if !response.Success {
		log.Fatal("API Request was unsuccessful.", log.NewAttr("message", response.Message))

		// Return to prevent further execution after log.Fatal().
		return
	}

	PrintCMDResponseFull(request, response, responseType, options, customPrintFunc)

	exit.Exit(0)
}

// Send a CMD request to the unix socket and return the response.
func SendCMDRequest(endpoint string, request any) (core.APIResponse, error) {
	socketPath, err := common.GetUnixSocketPath()
	if err != nil {
		return core.APIResponse{}, fmt.Errorf("Failed to get the unix socket path: '%w'.", err)
	}

	connection, err := net.Dial("unix", socketPath)
	if err != nil {
		return core.APIResponse{}, fmt.Errorf("Failed to dial the unix socket: '%w'.", err)
	}
	defer connection.Close()

	requestMap := map[string]any{
		server.ENDPOINT_KEY: endpoint,
		server.REQUEST_KEY:  request,
	}

	jsonRequest := util.MustToJSONIndent(requestMap)
	jsonBytes := []byte(jsonRequest)

	err = util.WriteToNetworkConnection(connection, jsonBytes)
	if err != nil {
		return core.APIResponse{}, fmt.Errorf("Failed to write the request to the unix socket: '%w'.", err)
	}

	responseBuffer, err := util.ReadFromNetworkConnection(connection)
	if err != nil {
		return core.APIResponse{}, fmt.Errorf("Failed to read the response from the unix socket: '%w'.", err)
	}

	var response core.APIResponse
	util.MustJSONFromBytes(responseBuffer, &response)

	return response, nil
}

func PrintCMDResponse(request any, response core.APIResponse, responseType any) {
	PrintCMDResponseFull(request, response, responseType, CommonOptions{}, nil)
}

// Print the CMD response in it's expected or custom format.
// customPrintFunc defines a function that takes the CMD response and formats it into a custom output.
// If verbose is true, it also displays the complete request and response.
func PrintCMDResponseFull(request any, response core.APIResponse, responseType any, options CommonOptions, customPrintFunc CustomResponseFormatter) {
	if options.Verbose {
		fmt.Printf("\nAutograder Request:\n---\n%s\n---\n", util.MustToJSONIndent(request))
		fmt.Printf("\nAutograder Response:\n---\n%s\n---\n", util.MustToJSONIndent(response))
	}

	successfulConversion := false
	output := ""
	if customPrintFunc != nil {
		output, successfulConversion = customPrintFunc(response)
	}

	if !successfulConversion {
		if responseType != nil {
			responseContent := reflect.New(reflect.TypeOf(responseType)).Interface()
			util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)
			output = util.MustToJSONIndent(responseContent)
		} else {
			output = util.MustToJSONIndent(response.Content)
		}
	}

	fmt.Println(output)
}

// Attempt to convert an API response into a TSV table.
// If the response content has one key with a list, the values in each element becomes a row of the table.
// If it has multiple keys or a single map, it creates a single row table.
// Return the table and true on a successful conversion or false if there is an issue converting.
func ConvertAPIResponseToTable(response core.APIResponse) (string, bool) {
	responseContent, ok := response.Content.(map[string]any)
	if !ok {
		return "", false
	}

	// Return false if the response content is empty.
	if len(responseContent) == 0 {
		return "", false
	}

	// Create a single row table if the response content has multiple keys.
	if len(responseContent) > 1 {
		return convertEntriesToTable([]any{responseContent})
	}

	// Retrieve the only key from the response content.
	var responseContentKey string
	for key := range responseContent {
		responseContentKey = key
	}

	var entries []any
	switch value := responseContent[responseContentKey].(type) {
	// Create a single row table if the response content has a single map.
	case map[string]any:
		entries = []any{value}
	// Create a multi-row table if the response content has a single list.
	case []any:
		entries = value
	default:
		return "", false
	}

	return convertEntriesToTable(entries)
}

// Attempt to convert a []may[string]any into a TSV table.
// Headers are taken from the keys of the first entry.
// Rows are the value's of each entry in the list.
// Return the table and true on a successful conversion or false if there is an issue converting.
func convertEntriesToTable(entries []any) (string, bool) {
	if len(entries) == 0 {
		return "", false
	}

	firstEntry, ok := entries[0].(map[string]any)
	if !ok {
		return "", false
	}

	var headers []string
	for key := range firstEntry {
		headers = append(headers, key)
	}

	sort.Strings(headers)

	var customTable strings.Builder
	customTable.WriteString(strings.Join(headers, "\t"))

	// Turn each entry into a row of the table.
	for _, entry := range entries {
		entryMap, ok := entry.(map[string]any)
		if !ok {
			return "", false
		}

		var row []string
		for _, key := range headers {
			switch value := entryMap[key].(type) {
			case map[string]any, []any:
				row = append(row, util.MustToJSON(value))
			default:
				row = append(row, fmt.Sprintf("%v", value))
			}
		}

		customTable.WriteString("\n" + strings.Join(row, "\t"))
	}

	return customTable.String(), true
}
