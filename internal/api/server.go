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
	"time"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var API_REQUEST_CONTENT_KEY = "content"
var AUTHENTICATION_NONCE = "nonce"
var port = config.WEB_PORT.Get()
var socketPath = config.UNIX_SOCKET_PATH.Get()
var pidPath = config.PID_PATH.Get()

// Run the standard API and Unix server.
func StartServer() error {
	serverErrorChannel := make(chan error, 2)

	go func() {
		serverErrorChannel <- startAPIServer()
	}()

	time.Sleep(10 * time.Millisecond)

	go func() {
		serverErrorChannel <- startUnixServer()
	}()

	for {
		select {
		case err := <-serverErrorChannel:
			return fmt.Errorf("server error: %v", err)
		case <-waitForShutdownSignal():
			fmt.Println("Shutting down the unix and api server")
			return nil
		}
	}
}

func waitForShutdownSignal() <-chan os.Signal {
	serverShutdownChannel := make(chan os.Signal, 1)

	signal.Notify(serverShutdownChannel, os.Interrupt, syscall.SIGTERM)

	return serverShutdownChannel
}

func StartUnixServer() {
	defer func() {
		err := os.Remove(socketPath)
		if err != nil {
			log.Error("Failed to remove the unix socket file path.", err)
		}
	}()

	err := startUnixServer()
	if err != nil {
		log.Fatal("Failed to start the Unix server.", err)
	}
}

func startAPIServer() error {
	defer func() {
		err := os.Remove(pidPath)
		if err != nil {
			log.Error("Failed to remove the PID file path.", err)
		}
	}()

	log.Info("API Server Started", log.NewAttr("port", port))
	return http.ListenAndServe(fmt.Sprintf(":%d", port), core.GetRouteServer(GetRoutes()))
}

func startUnixServer() error {
	unixListener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal("Failed to listen on a Unix socket.", err)
	}

	defer unixListener.Close()

	log.Info("Unix Server Started", log.NewAttr("unix_socket", socketPath))

	for {
		connection, err := unixListener.Accept()
		if err != nil {
			log.Fatal("Failed to accept a unix connection.", err)
		}

		go handleConnection(connection)
	}
}

func handleConnection(conn net.Conn) {
	sizeBuffer := make([]byte, 8)
	_, err := conn.Read(sizeBuffer)
	if err != nil {
		log.Fatal("Failed to read size of the request buffer.", err)
	}

	size := binary.BigEndian.Uint64(sizeBuffer)

	jsonBuffer := make([]byte, size)
	_, err = conn.Read(jsonBuffer)
	if err != nil {
		log.Error("Failed to read the request from the buffer.", err)
	}

	randomNumber, err := util.RandHex(64)
	if err != nil {
		log.Error("Failed to generate the nonce.", err)
	}
	core.NonceMap.Store(randomNumber, true)
	defer core.NonceMap.Delete(randomNumber)

	// Unmarshal the received JSON buffer into a map.
	var payload map[string]interface{}
	err = json.Unmarshal(jsonBuffer, &payload)
	if err != nil {
		log.Error("Failed to unmarshal the request buffer into the payload.", err)
	}

	endpoint, exists := payload["endpoint"].(string)
	if !exists {
		log.Error("Failed to find the 'endpoint' key in the request", exists)
	}

	content, exists := payload["request"].(map[string]interface{})
	if !exists {
		log.Error("Failed to find the 'request' key in the request.", exists)
	}

	content["root-user-nonce"] = randomNumber

	// Convert the modified content with the nonce back to JSON bytes.
	formContent, err := json.Marshal(content)
	if err != nil {
		log.Error("Failed to marshal the request's content.", err)
	}

	form := make(map[string]string)
	form[API_REQUEST_CONTENT_KEY] = string(formContent)

	url := "http://127.0.0.1" + fmt.Sprintf(":%d", port) + endpoint
	responseText, err := common.PostNoCheck(url, form)
	if err != nil {
		log.Error("Failed to POST an API request.", err)
	}

	jsonResponseBytes := []byte(responseText)
	size = uint64(len(jsonResponseBytes))
	responseBuffer := new(bytes.Buffer)

	err = binary.Write(responseBuffer, binary.BigEndian, size)
	if err != nil {
		log.Error("Failed to write response size to response buffer.", err)
	}

	responseBuffer.Write(jsonResponseBytes)

	_, err = conn.Write(responseBuffer.Bytes())
	if err != nil {
		log.Fatal("Failed to write the request back to the client.", err)
	}
}
