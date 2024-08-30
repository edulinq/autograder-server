package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/assignments/submissions"
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
	TargetSubmission string `help:"ID of the submission." arg:"" optional:""`
	Verbose          bool   `help:"Print the entire response." short:"v"`
}

const (
	BUFFER_SIZE = 8
)

func main() {
	kong.Parse(&args,
		kong.Description("Fetch a submission for a specific assignment and user."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Failed to load config options.", err)
	}

	socketPath, err := common.WriteAndReturnUnixSocketPath()
	if err != nil {
		log.Fatal("Failed to get unix socket path.", err)
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		log.Fatal("Failed to dial the unix socket.", err)
	}

	defer conn.Close()

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

	requestMap := map[string]interface{}{
		"endpoint": core.NewEndpoint(`courses/assignments/submissions/fetch/user/peek`),
		"request":  request,
	}

	jsonRequest := util.MustToJSONIndent(requestMap)
	jsonBytes := []byte(jsonRequest)
	requestBuffer := new(bytes.Buffer)
	size := uint64(len(jsonBytes))

	err = binary.Write(requestBuffer, binary.BigEndian, size)
	if err != nil {
		log.Fatal("Failed to write message size to the request buffer.", err)
	}

	requestBuffer.Write(jsonBytes)

	_, err = conn.Write(requestBuffer.Bytes())
	if err != nil {
		log.Fatal("Failed to write the request buffer to the unix server.", err)
	}

	sizeBuffer := make([]byte, BUFFER_SIZE)

	_, err = conn.Read(sizeBuffer)
	if err != nil {
		log.Fatal("Failed to read the size of the response buffer.", err)
	}

	size = binary.BigEndian.Uint64(sizeBuffer)
	responseBuffer := make([]byte, size)

	_, err = conn.Read(responseBuffer)
	if err != nil {
		log.Fatal("Failed to read the response.", err)
	}

	var response core.APIResponse
	err = json.Unmarshal(responseBuffer, &response)
	if err != nil {
		log.Fatal("Failed to unmarshal the API response.", err)
	}

	if !response.Success {
		log.Fatal("Failed to make an api request. Status: '%d' ", response.HTTPStatus)
	}

	if args.Verbose {
		fmt.Println(util.MustToJSONIndent(response))
	} else {
		var responseContent submissions.FetchUserPeekResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)
		fmt.Println(util.MustToJSONIndent(responseContent))
	}
}
