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

func SendAndPrintCMDRequest(endpoint string, request any, responseType any, verbose bool) error {
	response, err := SendCMDRequest(endpoint, request)
	if err != nil {
		return fmt.Errorf("Failed to send the CMD request: '%w'.", err)
	}

	PrintCMDResponse(request, response, responseType, verbose)

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

// Print the CMD response in it's expected JSON format.
// If verbose is true, it also displays the complete request and response.
func PrintCMDResponse(request any, response core.APIResponse, responseType any, verbose bool) {
	if verbose {
		fmt.Printf("\nAutograder Request:\n---\n%s\n---\n", util.MustToJSONIndent(request))
		fmt.Printf("\nAutograder Response:\n---\n%s\n---\n", util.MustToJSONIndent(response))
	}

	if !response.Success {
		fmt.Println(response.Message)

		util.Exit(2)
		return
	}

	responseContent := reflect.New(reflect.TypeOf(responseType)).Interface()
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)
	fmt.Println(util.MustToJSONIndent(responseContent))
}
