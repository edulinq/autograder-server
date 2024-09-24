package cmd

import (
	"fmt"
	"net"
	"reflect"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/server"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/util"
)

func SendAndPrintCMDRequest(endpoint string, request any, responseType any, verbose bool) error {
	response, err := SendCMDRequest(endpoint, request)
	if err != nil {
		return fmt.Errorf("Failed to send the CMD request: '%w'.", err)
	}

	if verbose {
		PrintVerboseCMDResponse(request, response, responseType)
	} else {
		PrintCMDResponse(response, responseType)
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

// Print the CMD response in it's expected format.
// If in testing mode, print the CMD as a core.APIResponse type.
func PrintCMDResponse(response core.APIResponse, responseType any) {
	testingMode := config.TESTING_MODE.Get()

	if !response.Success {
		if testingMode {
			fmt.Println(util.MustToJSONIndent(response))
		} else {
			fmt.Println(response.Message)
		}

		util.Exit(2)
		return
	}

	if testingMode {
		fmt.Println(util.MustToJSONIndent(response))
		return
	}

	responseContent := reflect.New(reflect.TypeOf(responseType)).Interface()
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)
	fmt.Println(util.MustToJSONIndent(responseContent))
}

func PrintVerboseCMDResponse(request any, response core.APIResponse, responseType any) {
	fmt.Println("Request:")
	fmt.Println(util.MustToJSONIndent(request))
	fmt.Println("Response:")
	fmt.Println(util.MustToJSONIndent(response))

	if !response.Success {
		util.Exit(2)
		return
	}

	responseContent := reflect.New(reflect.TypeOf(responseType)).Interface()
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)
	fmt.Println(util.MustToJSONIndent(responseContent))
}
