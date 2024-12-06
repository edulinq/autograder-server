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
			defer server.StopServer()
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
	customPrintOutput := ""
	if customPrintFunc != nil {
		customPrintOutput, successfulConversion = customPrintFunc(response)
	}

	if successfulConversion && customPrintFunc != nil {
		fmt.Println(customPrintOutput)
	} else if responseType == nil {
		fmt.Println(util.MustToJSONIndent(response.Content))
	} else {
		responseContent := reflect.New(reflect.TypeOf(responseType)).Interface()
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)
		fmt.Println(util.MustToJSONIndent(responseContent))
	}
}

// Attempt to convert an API response to a TSV table.
// Return ("", false) if there are issues converting the API response
// or (customTable.String(), true) if the response got sucessfully converted.
func AttemptApiResponseToTable(response core.APIResponse) (string, bool) {
	responseContent, ok := response.Content.(map[string]any)
	if !ok {
		return "", false
	}

	// Don't try to convert a response that has multiple keys.
	if len(responseContent) != 1 {
		return "", false
	}

	responseContentKey := ""
	for key := range responseContent {
		responseContentKey = key
	}

	// Get the rows that will be added to the table.
	entries, ok := responseContent[responseContentKey].([]any)
	if !ok {
		return "", false
	}

	if len(entries) == 0 {
		return "", false
	}

	// Use the first entry to create the headers of the table.
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
	customTable.WriteString(strings.Join(headers, "\t") + "\n")

	// Turn each entry into a row of the table.
	for i, entry := range entries {
		entryMap, ok := entry.(map[string]any)
		if !ok {
			return "", false
		}

		var row []string
		for _, key := range headers {
			switch value := entryMap[key].(type) {
			case map[string]any:
				row = append(row, util.MustToJSON(value))
			default:
				row = append(row, fmt.Sprintf("%v", value))
			}
		}

		customTable.WriteString(strings.Join(row, "\t"))

		if i < (len(entries) - 1) {
			customTable.WriteString("\n")
		}
	}

	return customTable.String(), true
}
