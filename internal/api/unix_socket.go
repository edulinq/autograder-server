package api

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

func startExclusiveUnixServer() error {
	var socketPath = config.UNIX_SOCKET_PATH.Get()

	unixSocket, err := net.Listen("unix", socketPath)
	defer os.Remove(socketPath)
	if err != nil {
		log.Fatal("Failed to listen on a Unix socket.", err)
	}

	// Remove the unix socket file when the program terminates abruptly.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove(socketPath)
		os.Exit(1)
	}()

	defer unixSocket.Close()

	log.Info("Unix Server Started", log.NewAttr("unix_socket", socketPath))

	for {
		connection, err := unixSocket.Accept()
		if err != nil {
			log.Error("Failed to accept a unix connection.", err)
		}

		go handleUnixSocketConnection(connection)
	}
}

func handleUnixSocketConnection(conn net.Conn) {
	var port = config.WEB_PORT.Get()
	var bufferBytes = config.BUFFER_SIZE.Get()
	var bitSize = config.BIT_SIZE.Get()

	jsonBuffer, err := util.ReadFromUnixSocket(conn, bufferBytes)
	if err != nil {
		log.Error("Failed to read from UNIX socket", err)
		return
	}

	randomNumber, err := util.RandHex(bitSize)
	if err != nil {
		log.Error("Failed to generate the nonce.", err)
		return
	}
	core.RootUserNonces.Store(randomNumber, true)
	defer core.RootUserNonces.Delete(randomNumber)

	var payload map[string]any
	err = json.Unmarshal(jsonBuffer, &payload)
	if err != nil {
		log.Error("Failed to unmarshal the request buffer into the payload.", err)
		return
	}

	endpoint, exists := payload["endpoint"].(string)
	if !exists {
		log.Error("Failed to find the 'endpoint' key in the request", exists)
		return
	}

	content, exists := payload["request"].(map[string]any)
	if !exists {
		log.Error("Failed to find the 'request' key in the request.", exists)
		return
	}

	content["root-user-nonce"] = randomNumber
	formContent, err := json.Marshal(content)
	if err != nil {
		log.Error("Failed to marshal the request's content.", err)
		return
	}

	form := make(map[string]string)
	form[API_REQUEST_CONTENT_KEY] = string(formContent)

	url := fmt.Sprintf("http://127.0.0.1:%d%s", port, endpoint)
	responseText, err := common.PostNoCheck(url, form)
	if err != nil {
		log.Error("Failed to POST an API request.", err)
		return
	}

	jsonResponseBytes := []byte(responseText)

	err = util.WriteToUnixSocket(conn, jsonResponseBytes)
	if err != nil {
		log.Error("Failed to write response to UNIX socket", err)
	}

}
