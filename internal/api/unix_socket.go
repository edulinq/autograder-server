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

const NONCE_BYTE_SIZE = 64

func runUnixSocketServer() (err error) {
	defer func() {
		value := recover()
		if value == nil {
			return
		}

		err = fmt.Errorf("Unix socket panicked: '%v'.", value)
	}()

	unixSocketPath, err := common.WriteAndReturnUnixSocketPath()
	if err != nil {
		return err
	}

	util.MkDir(filepath.Dir(unixSocketPath))

	unixSocket, err = net.Listen("unix", unixSocketPath)
	if err != nil {
		log.Error("Failed to listen on a Unix socket.", err)
		return err
	}
	defer StopUnixSocketServer()

	log.Info("Unix Socket Server Started", log.NewAttr("unix_socket", unixSocketPath))

	for {
		connection, err := unixSocket.Accept()
		if err != nil {
			log.Info("Unix Socket Server Stopped", log.NewAttr("unix_socket", unixSocketPath))

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

	jsonBuffer, err := util.ReadFromUnixSocket(conn)
	if err != nil {
		return fmt.Errorf("Failed to read from the unix socket.")
	}

	randomNumber, err := util.RandHex(NONCE_BYTE_SIZE)
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
		return fmt.Errorf("Failed to write to the unix socket.")
	}

	return nil
}
