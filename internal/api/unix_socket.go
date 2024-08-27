package api

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
	BIT_SIZE    = 64
	BUFFER_SIZE = 8
)

func runUnixSocketServer() (err error) {
	defer func() {
		value := recover()
		if value == nil {
			return
		}

		err = fmt.Errorf("Unix socket panicked: '%v'.", value)
	}()

	var socketPath = config.GetUnixSocketDir()
	util.MkDir(filepath.Dir(socketPath))

	unixSocket, err = net.Listen("unix", socketPath)
	if err != nil {
		log.Error("Failed to listen on a Unix socket.", err)
		return err
	}

	defer StopUnixSocketServer()

	log.Info("Unix Socket Server Started", log.NewAttr("unix_socket", socketPath))

	for {
		connection, err := unixSocket.Accept()
		if err != nil {
			log.Info("Unix Socket Server Stopped", log.NewAttr("unix_socket", socketPath))

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
				log.Error("Error handling Unix socket connection.", err)
			}
		}()
	}
}

func handleUnixSocketConnection(conn net.Conn) error {
	var port = config.WEB_PORT.Get()
	var bufferBytes = BUFFER_SIZE
	var bitSize = BIT_SIZE

	jsonBuffer, err := util.ReadFromUnixSocket(conn, bufferBytes)
	if err != nil {
		return fmt.Errorf("Failed to read from UNIX socket.")
	}

	randomNumber, err := util.RandHex(bitSize)
	if err != nil {
		return fmt.Errorf("Failed to generate the nonce.")
	}

	core.RootUserNonces.Store(randomNumber, true)
	defer core.RootUserNonces.Delete(randomNumber)

	var payload map[string]any
	err = json.Unmarshal(jsonBuffer, &payload)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal the request buffer into the payload.")
	}

	endpoint, exists := payload["endpoint"].(string)
	if !exists {
		return fmt.Errorf("Failed to find the 'endpoint' key in the request.")
	}

	content, exists := payload["request"].(map[string]any)
	if !exists {
		return fmt.Errorf("Failed to find the 'request' key in the request.")
	}

	content["root-user-nonce"] = randomNumber
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

	err = util.WriteToUnixSocket(conn, jsonResponseBytes)
	if err != nil {
		return err
	}

	return nil
}
