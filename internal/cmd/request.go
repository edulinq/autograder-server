package cmd

import (
	"fmt"
	"net"
	"reflect"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/server"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/util"
)

type CommonCMDArgs struct {
	Verbose bool `help:"Use verbose output to show full request/response without specific formatting." default:"false"`
}

func SendAndPrintCMDRequest(endpoint string, request any, responseType any, verbose bool, customPrint func()) error {
	response, err := SendCMDRequest(endpoint, request)
	if err != nil {
		return fmt.Errorf("Failed to send the CMD request: '%w'.", err)
	}

	if customPrint != nil {
		PrintCMDResponseFull(request, response, responseType, verbose, customPrint)
	} else {
		PrintCMDResponseFull(request, response, responseType, verbose, nil)
	}

	return nil
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

func PrintCMDResponse(request any, response core.APIResponse, responseType any) {
	PrintCMDResponseFull(request, response, responseType, false, nil)
}

// Print the CMD response in it's expected JSON format.
// If verbose is true, it also displays the complete request and response.
func PrintCMDResponseFull(request any, response core.APIResponse, responseType any, verbose bool, customPrint func()) {
	if !response.Success {
		fmt.Println(response.Message)

		util.Exit(2)
		return
	}

	if verbose {
		fmt.Printf("\nAutograder Request:\n---\n%s\n---\n", util.MustToJSONIndent(request))
		fmt.Printf("\nAutograder Response:\n---\n%s\n---\n", util.MustToJSONIndent(response))
	}

	if customPrint != nil {
		customPrint()
	} else {
		responseContent := reflect.New(reflect.TypeOf(responseType)).Interface()
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)
		fmt.Println(util.MustToJSONIndent(responseContent))
	}
}
