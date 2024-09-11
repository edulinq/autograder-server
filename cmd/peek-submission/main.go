package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/assignments/submissions"
	"github.com/edulinq/autograder/internal/api/server"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
	TargetEmail      string `help:"Email of the user to fetch." arg:""`
	CourseID         string `help:"ID of the course." arg:""`
	AssignmentID     string `help:"ID of the assignment." arg:""`
	TargetSubmission string `help:"ID of the submission. Defaults to the latest submission." arg:"" optional:""`
	ShortForm        bool   `help:"Use short form output." long:"short-form"`
}

func main() {
	kong.Parse(&args,
		kong.Description("Fetch a submission for a specific assignment and user."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Failed to load config options.", err)
	}

	socketPath, err := common.GetUnixSocketPath()
	if err != nil {
		log.Fatal("Failed to get the unix socket path.", err)
	}

	connection, err := net.Dial("unix", socketPath)
	if err != nil {
		log.Fatal("Failed to dial the unix socket.", err)
	}
	defer connection.Close()

	request := submissions.FetchUserPeekRequest{
		APIRequestAssignmentContext: core.APIRequestAssignmentContext{
			APIRequestCourseUserContext: core.APIRequestCourseUserContext{
				CourseID: args.CourseID,
			},
			AssignmentID: args.AssignmentID,
		},
		TargetUser: core.TargetCourseUserSelfOrGrader{
			TargetCourseUser: core.TargetCourseUser{
				Email: args.TargetEmail,
			},
		},
		TargetSubmission: args.TargetSubmission,
	}

	requestMap := map[string]any{
		server.ENDPOINT_KEY: core.NewEndpoint(`courses/assignments/submissions/fetch/user/peek`),
		server.REQUEST_KEY:  request,
	}

	jsonRequest := util.MustToJSONIndent(requestMap)
	jsonBytes := []byte(jsonRequest)
	err = util.WriteToNetworkConnection(connection, jsonBytes)
	if err != nil {
		log.Fatal("Failed to write the request to the unix socket.", err)
	}

	responseBuffer, err := util.ReadFromNetworkConnection(connection)
	if err != nil {
		log.Fatal("Failed to read the response from the unix socket.", err)
	}

	var response core.APIResponse
	err = json.Unmarshal(responseBuffer, &response)
	if err != nil {
		log.Fatal("Failed to unmarshal the API response.", err)
	}

	if !response.Success {
		output := response.Message
		if !args.ShortForm {
			output = util.MustToJSONIndent(response)
		}

		fmt.Println(output)

		os.Exit(2)
	}

	if args.ShortForm {
		var responseContent submissions.FetchUserPeekResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)
		fmt.Println(util.MustToJSONIndent(responseContent))
	} else {
		fmt.Println(util.MustToJSONIndent(response))
	}
}
