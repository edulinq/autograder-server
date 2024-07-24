package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api"
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


}

func main() {
	kong.Parse(&args,
		kong.Description("Fetch the submission for a specific assignment and user."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	go func() {
		api.StartUnixServer()
	}()

	time.Sleep(1 * time.Second)

	socketPath := config.UNIX_SOCKET.Get();

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
	request := submissions.PeekRequest{
		TargetUser:           targetUser,
		TargetSubmission:     args.TargetSubmission,
	}
	jsonRequest := util.MustToJSONIndent(request)
	jsonBytes := []byte(jsonRequest)
	buffer := new(bytes.Buffer)
	size := uint64(len(jsonBytes))

	err = binary.Write(buffer, binary.BigEndian, size)
	if err != nil {
		log.Fatal("Failed to write message size to buffer.", err)
	}
	// fmt.Println("buffer with size: ", buffer.Bytes())
	buffer.Write(jsonBytes)
	// fmt.Println("Buffer with size + json", buffer.Bytes())

	_, err = conn.Write(buffer.Bytes())
	if (err != nil) {
		log.Fatal("Failed to send request to the server.", err)
	}
	responseBuffer := make([]byte, 4096)
	response, err := conn.Read(responseBuffer)
	if (err != nil) {
		log.Fatal("Failed to read response.", err)
	}
	fmt.Println("response client: ", string(responseBuffer[:response]))
}

// func sendRequest(conn net.Conn, request map[string]string) {
// 	jsonRequest, err := json.Marshal(request)
// 	if err != nil {
// 		log.Fatal("Failed to marshal request.", err)
// 	}
	

// 	buffer := new(bytes.Buffer)
// 	size := uint64(len(jsonRequest))
// 	err = binary.Write(buffer, binary.BigEndian, size)
// 	if err != nil {
// 		log.Fatal("Failed to write message size to buffer.", err)
// 	}

// 	buffer.Write(jsonRequest)
// 	fmt.Println("buffer: ", buffer)
// 	_, err = conn.Write(buffer.Bytes())
// 	if err != nil {
// 		log.Fatal("Failed to send request to the server.", err)
// 	}
// }

// func readResponse(conn net.Conn) int{

// 	// Read the response based on the size
// 	sizeBuffer := make([]byte, 8)
// 	_, err := conn.Read(sizeBuffer)
// 	if err != nil {
// 		log.Fatal("Failed to read size response.", err)
// 	}

// 	size := binary.BigEndian.Uint64(sizeBuffer)


// 	randNumBuffer := make([]byte, size)
// 	_, err = conn.Read(randNumBuffer)
// 	if err != nil {
// 		log.Error("Failed to read random number response.", err)
// 	}

// 	randomNumber := new(big.Int).SetBytes(randNumBuffer)

// 	return int(randomNumber.Int64())
// }