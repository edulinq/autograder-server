package cmd

import (
	"fmt"
	"net"
	"reflect"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/server"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

type CommonOptions struct {
	Verbose bool `help:"Add the full request and response to the output. Note: the output will include extra details beyond the expected format." default:"false"`
}

type CustomFormat func(response core.APIResponse) string

func MustHandleCMDRequestAndExit(endpoint string, request any, responseType any) {
	MustHandleCMDRequestAndExitFull(endpoint, request, responseType, CommonOptions{}, nil)
}

func MustHandleCMDRequestAndExitFull(endpoint string, request any, responseType any, options CommonOptions, customPrintFunc CustomFormat) {
	response, err := SendCMDRequest(endpoint, request)
	if err != nil {
		log.Fatal("Failed to send the CMD request", err)
	}

	exitCode := PrintCMDResponseFull(request, response, responseType, options, customPrintFunc)

	util.Exit(exitCode)
}

// Send a CMD request to the unix socket and return the response.
func SendCMDRequest(endpoint string, request any) (core.APIResponse, error) {
	socketPath, err := common.GetUnixSocketPath()
	if err != nil {
		return core.APIResponse{}, fmt.Errorf("Failed to get the unix socket path: '%w'.", err)
	}

	connection, err := net.Dial("unix", socketPath)
	if err != nil {
		return core.APIResponse{}, fmt.Errorf("Failed to dial the unix socket: '%w'", err)
	}

	defer connection.Close()

	requestMap := map[string]any{
		server.ENDPOINT_KEY: core.NewEndpoint(endpoint),
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
	if err != nil {
		return core.APIResponse{}, fmt.Errorf("Failed to unmarshal the API response: '%w'.", err)
	}

	return response, nil
}

func PrintCMDResponse(request any, response core.APIResponse, responseType any) int {
	return PrintCMDResponseFull(request, response, responseType, CommonOptions{}, nil)
}

// Print the CMD response in it's expected or custom format.
// If verbose is true, it also displays the complete request and response.
func PrintCMDResponseFull(request any, response core.APIResponse, responseType any, options CommonOptions, customPrintFunc CustomFormat) int {
	if options.Verbose {
		fmt.Printf("\nAutograder Request:\n---\n%s\n---\n", util.MustToJSONIndent(request))
		fmt.Printf("\nAutograder Response:\n---\n%s\n---\n", util.MustToJSONIndent(response))
	}

	if !response.Success {
		log.Error("API response was unsuccessful.", log.NewAttr("message", response.Message))
		return 2
	}

	if customPrintFunc != nil {
		fmt.Println(customPrintFunc(response))
	} else {
		responseContent := reflect.New(reflect.TypeOf(responseType)).Interface()
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)
		fmt.Println(util.MustToJSONIndent(responseContent))
	}

	return 0
}
