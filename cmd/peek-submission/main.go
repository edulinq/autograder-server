package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/submissions"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
	TargetEmail      string `help:"Email of the user to fetch." arg:""`
	TargetSubmission string `help:"ID of the submission." arg:""`

	CourseID     string `help:"ID of the course." arg:""`
	AssignmentID string `help:"ID of the assignment." arg:""`
}

func main() {
	kong.Parse(&args,
		kong.Description("Fetch the submission for a specific assignment and user."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Failed to load config options.", err)
	}

	socketPath := config.UNIX_SOCKET_PATH.Get()

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		log.Fatal("Failed to dial the unix socket.", err)
		os.Exit(1)
	}

	defer conn.Close()

	targetUser := core.TargetCourseUserSelfOrGrader{
		TargetCourseUser: core.TargetCourseUser{
			Email: args.TargetEmail,
		},
	}

	assignmentContext := core.APIRequestAssignmentContext{
		APIRequestCourseUserContext: core.APIRequestCourseUserContext{
			CourseID: args.CourseID,
		},
		AssignmentID: args.AssignmentID,
	}

	request := map[string]interface{}{
		"endpoint": core.NewEndpoint(`submissions/peek`),
		"request": submissions.PeekRequest{
			APIRequestAssignmentContext: assignmentContext,
			TargetUser:                  targetUser,
			TargetSubmission:            args.TargetSubmission,
		},
	}

	jsonRequest := util.MustToJSONIndent(request)
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

	sizeBuffer := make([]byte, 8)
	_, err = conn.Read(sizeBuffer)
	if err != nil {
		log.Error("Failed to read the size of the response buffer.", err)
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
		log.Error("Failed to unmarshal the API response.", err)
		return
	}

	var responseContent submissions.PeekResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	fmt.Println(util.MustToJSONIndent(responseContent))
}
