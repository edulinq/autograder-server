package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/submissions"
	"github.com/edulinq/autograder/internal/util"

	// "github.com/edulinq/autograder/internal/api/core"
	// "github.com/edulinq/autograder/internal/api/submissions"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	// "github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
	TargetEmail      string `help:"Email of the user to fetch." arg:""`
	TargetSubmission string `help:"ID of the submission." arg:""`

	CourseID         string `help:"ID of the course." arg:""`
	AssignmentID     string `help:"ID of the assignment." arg:""`

}

func main() {
	kong.Parse(&args,
		kong.Description("Fetch the submission for a specific assignment and user."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	socketPath := config.UNIX_SOCKET_PATH.Get();

	conn, err := net.Dial("unix", socketPath)
	if (err != nil) {
		log.Fatal("Failed to dial the unix socket. ", err)
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
		"endpoint":                 core.NewEndpoint(`submissions/peek`),
		"request":                  submissions.PeekRequest{
			APIRequestAssignmentContext: assignmentContext,
			TargetUser:       targetUser,
			TargetSubmission: args.TargetSubmission,
		},
	}
	// request := submissions.PeekRequest{
	// 	APIRequestAssignmentContext: assignmentContext,
	// 	TargetUser:       targetUser,
	// 	TargetSubmission: args.TargetSubmission,
	// }

	jsonRequest := util.MustToJSONIndent(request)
	jsonBytes := []byte(jsonRequest)
	buffer := new(bytes.Buffer)
	size := uint64(len(jsonBytes))

	err = binary.Write(buffer, binary.BigEndian, size)
	if err != nil {
		log.Fatal("Failed to write message size to buffer.", err)
	}

	buffer.Write(jsonBytes)

	_, err = conn.Write(buffer.Bytes())
	if (err != nil) {
		log.Fatal("Failed to send request to the server.", err)
	}

	sizeBuffer := make([]byte, 8)
	_, err = conn.Read(sizeBuffer)
	if err != nil {
		log.Error("Failed to read the size of the payload.", err)
	}

	size = binary.BigEndian.Uint64(sizeBuffer)

	responseBuffer := make([]byte, size)
	response, err := conn.Read(responseBuffer)
	if (err != nil) {
		log.Fatal("Failed to read response.", err)
	}
	fmt.Println("response client: ", string(responseBuffer[:response]))
}
