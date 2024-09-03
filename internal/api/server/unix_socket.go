package server

import (
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

const (
	UNIX_SOCKET_SERVER_STOP_LOCK = "UNIX_SOCKET_STOP_LOCK"
	API_REQUEST_CONTENT_KEY      = "content"
	NONCE_SIZE_BYTES             = 64
	ENDPOINT_KEY                 = "endpoint"
	REQUEST_KEY                  = "request"
	NONCE_KEY                    = "root-user-nonce"
)

var unixSocket net.Listener

func RunUnixSocketServer() (err error) {
	defer func() {
		value := recover()
		if value == nil {
			return
		}

		err = fmt.Errorf("Unix socket panicked: '%v'.", value)
	}()

	unixSocketPath, err := common.GetUnixSocketPath()
	if err != nil {
		return err
	}

	util.MkDir(filepath.Dir(unixSocketPath))

	unixSocket, err = net.Listen("unix", unixSocketPath)
	if err != nil {
		log.Error("Failed to listen on the unix socket.", err)
		return err
	}
	defer StopUnixSocketServer()

	log.Info("Unix Socket Server Started.", log.NewAttr("unix_socket", unixSocketPath))

	for {
		connection, err := unixSocket.Accept()
		if err != nil {
			log.Info("Unix Socket Server Stopped.", log.NewAttr("unix_socket", unixSocketPath))

			if unixSocket == nil {
				return nil
			}

			log.Error("Unix socket server returned an error.", err)
			return err
		}

		go func() {
			defer connection.Close()

			err := handleUnixSocketConnection(connection)
			if err != nil {
				log.Error("Error handling the unix socket connection.", err)
			}
		}()
	}
}

// Read the request from the unix socket, generate and store a nonce, add the nonce to the request,
// send the request to the API endpoint, and write the response back to the unix socket.
func handleUnixSocketConnection(connection net.Conn) error {
	var port = config.WEB_PORT.Get()

	jsonBuffer, err := util.ReadFromUnixSocket(connection)
	if err != nil {
		return fmt.Errorf("Failed to read from the unix socket.")
	}

	randomNonce, err := util.RandHex(NONCE_SIZE_BYTES)
	if err != nil {
		return fmt.Errorf("Failed to generate the nonce.")
	}

	core.RootUserNonces.Store(randomNonce, true)
	defer core.RootUserNonces.Delete(randomNonce)

	var payload map[string]any
	err = json.Unmarshal(jsonBuffer, &payload)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal the request buffer into the payload.")
	}

	endpoint, exists := payload[ENDPOINT_KEY].(string)
	if !exists {
		return fmt.Errorf("Failed to find the 'endpoint' key in the request.")
	}

	content, exists := payload[REQUEST_KEY].(map[string]any)
	if !exists {
		return fmt.Errorf("Failed to find the 'request' key in the request.")
	}

	content[NONCE_KEY] = randomNonce
	formContent, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("Failed to marshal the request's content.")
	}

	form := make(map[string]string)
	form[API_REQUEST_CONTENT_KEY] = string(formContent)

	url := fmt.Sprintf("http://127.0.0.1:%d%s", port, endpoint)
	responseText, err := common.PostNoCheck(url, form)
	if err != nil {
		return fmt.Errorf("Failed to POST an API request.")
	}

	jsonResponseBytes := []byte(responseText)
	err = util.WriteToUnixSocket(connection, jsonResponseBytes)
	if err != nil {
		return fmt.Errorf("Failed to write to the unix socket.")
	}

	return nil
}

func StopUnixSocketServer() {
	common.Lock(UNIX_SOCKET_SERVER_STOP_LOCK)
	defer common.Unlock(UNIX_SOCKET_SERVER_STOP_LOCK)

	if unixSocket == nil {
		return
	}

	tempUnixSocket := unixSocket
	unixSocket = nil

	err := tempUnixSocket.Close()
	if err != nil {
		log.Fatal("Failed to close the unix socket.", err)
	}
}
