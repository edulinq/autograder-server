package api

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	// "time"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/util"

	// "github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	// "github.com/edulinq/autograder/internal/util"
)
var API_REQUEST_CONTENT_KEY = "content"
var AUTHENTICATION_NONCE = "nonce"
var port = config.WEB_PORT.Get()
var socketPath = config.UNIX_SOCKET_PATH.Get();
var pidPath = config.PID_PATH.Get();



// Run the standard API and Unix server.
func StartServer() error {
	serverErrorChannel := make(chan error, 2)

	go func() {
		serverErrorChannel <- startAPIServer()
	}()

	go func() {
		serverErrorChannel <- startUnixServer()
	}()

	for {
		select {
			case err := <-serverErrorChannel:
				return fmt.Errorf("server error: %v", err)
			case <- waitForShutdownSignal():
				return nil
    	}
	}
}

func waitForShutdownSignal() <-chan os.Signal {
    serverShutdownChannel := make(chan os.Signal, 1)
    signal.Notify(serverShutdownChannel, os.Interrupt, syscall.SIGTERM)
	os.Remove(socketPath)
	// os.Remove(pidPath)

    return serverShutdownChannel
}

func StartUnixServer() {
	err := startUnixServer()
	if (err != nil) {
		log.Fatal("Failed to start the Unix server.", err)
	}
}

func startAPIServer() error {

	log.Info("API Server Started", log.NewAttr("port", port))
	return http.ListenAndServe(fmt.Sprintf(":%d", port), core.GetRouteServer(GetRoutes()));
}

func startUnixServer() error {
	unixListener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal("Error creating Unix socket. ", err)
	}

	defer unixListener.Close()

	log.Info("Unix Server Started", log.NewAttr("unix_socket", socketPath))
	defer os.Remove(socketPath)
	
	for {
		connection, err := unixListener.Accept()
		if (err != nil) {
			log.Fatal("Failed to accept a connection. ", err)
		}
		go handleConnection(connection)
	}
}

func handleConnection(conn net.Conn) {	
	sizeBuffer := make([]byte, 8)
	_, err := conn.Read(sizeBuffer)
	if err != nil {
		log.Fatal("Failed to read size of the payload.", err)
	}

	size := binary.BigEndian.Uint64(sizeBuffer)
	
	jsonBuffer := make([]byte, size)
	_, err = conn.Read(jsonBuffer)
	if err != nil {
		log.Error("Failed to read JSON payload", err)
		return
	}

	randomNumber, err := util.RandHex(64)
	if err != nil {
		log.Error("Failed to generate random number.", err)
	}


	var payload map[string]interface{}
	err = json.Unmarshal(jsonBuffer, &payload)
	if err != nil {
		log.Error("Failed to unmarshal the JSON buffer.", err)
		return
	}

	endpoint, ok := payload["endpoint"].(string)
	if !ok {
		log.Error("Endpoint not found or invalid in the request")
	}


	content, ok := payload["request"].(map[string]interface{})
	if !ok {
		log.Error("Request content not found or invalid in the request")
	}

	content["root-user-nonce"] = randomNumber

	form := make(map[string]string)
	formContent, err := json.Marshal(content)
	if err != nil {
		log.Error("Failed to marshal form content to JSON.", err)
		return
	}
	form[API_REQUEST_CONTENT_KEY] = string(formContent)

	core.NonceMap.Store(randomNumber, true)
	defer core.NonceMap.Delete(randomNumber)

	url := "http://127.0.0.1" + fmt.Sprintf(":%d", port) + endpoint

	response, err := common.PostNoCheck(url, form)
	if (err != nil) {
		log.Error("Failed to POST an API request.", err)
	}

	jsonBytes := []byte(response)
	responseBuffer := new(bytes.Buffer)
	size = uint64(len(jsonBytes))

	err = binary.Write(responseBuffer, binary.BigEndian, size)
	if err != nil {
		log.Error("Failed to write response size to buffer.", err)
	}

	responseBuffer.Write(jsonBytes)

	_, err = conn.Write(responseBuffer.Bytes())
	if err != nil {
		log.Fatal("Failed to echo back data.", err)
	}
}
