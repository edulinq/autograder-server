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
	Verbose bool `help:"Use verbose output to show the full request/response. Output may not adhere to a valid format/filetype." default:"false"`
}

type CMDPrintHandler func(response core.APIResponse) string

func MustHandleCMDRequestAndExit(endpoint string, request any, responseType any) {
	MustHandleCMDRequestAndExitFull(endpoint, request, responseType, CommonOptions{}, nil)
}

func MustHandleCMDRequestAndExitFull(endpoint string, request any, responseType any, options CommonOptions, customPrintFunc CMDPrintHandler) {
	response, err := SendCMDRequest(endpoint, request)
	if err != nil {
		log.Fatal("Failed to send the CMD request", err)
	}

	PrintCMDResponseFull(request, response, responseType, options, customPrintFunc)

	if response.Success {
		util.Exit(0)
	}
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
	PrintCMDResponseFull(request, response, responseType, CommonOptions{}, nil)
}

// Print the CMD response in it's expected or custom format.
// If verbose is true, it also displays the complete request and response.
func PrintCMDResponseFull(request any, response core.APIResponse, responseType any, options CommonOptions, customPrintFunc CMDPrintHandler) {
	if options.Verbose {
		fmt.Printf("\nAutograder Request:\n---\n%s\n---\n", util.MustToJSONIndent(request))
		fmt.Printf("\nAutograder Response:\n---\n%s\n---\n", util.MustToJSONIndent(response))
	}

	if !response.Success {
		log.Error("API response was unsuccessful.", log.NewAttr("message", response.Message))
		// fmt.Println(response.Message)
		util.Exit(2)
		return
	}

	if customPrintFunc != nil {

		response := customPrintFunc(response)

		// customPrintFuncValue := reflect.ValueOf(customPrintFunc)

		// responseValue := reflect.ValueOf(response)
		// responseSlice := []reflect.Value{responseValue}

		// response := customPrintFuncValue.Call(responseSlice)[0]
		fmt.Println(response)
	} else {
		responseContent := reflect.New(reflect.TypeOf(responseType)).Interface()
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)
		fmt.Println(util.MustToJSONIndent(responseContent))
	}
}
