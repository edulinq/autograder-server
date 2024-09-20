package cmd

import (
	"fmt"
	"net"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/assignments/submissions"
	"github.com/edulinq/autograder/internal/api/server"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/util"
)

func SendAndPrintCMDRequest(endpoint string, request any, responseType any, shortForm bool) error {
    response, err := sendCMDRequest(endpoint, request)
    if err != nil {
        return fmt.Errorf("Failed to send the CMD request: '%w'.", err)
    }

    printCMDResponse(response, responseType, shortForm)

    return nil
}


// Send a CMD request to the unix socket and return the response.
func sendCMDRequest(endpoint string, request any) (core.APIResponse, error) {
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

func printCMDResponse(response core.APIResponse, responseType any, shortForm bool) {
	if !response.Success {
		output := response.Message
		if !shortForm {
			output = util.MustToJSONIndent(response)
		}

		fmt.Println(output)

		util.Exit(2)
		return
	}

	if shortForm {
		var responseContent submissions.FetchUserPeekResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)
		fmt.Println(util.MustToJSONIndent(responseContent))
	} else {
		fmt.Println(util.MustToJSONIndent(response))
	}
}
