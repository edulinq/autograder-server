package api

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

func startExclusiveUnixServer() error {
	var socketPath = config.UNIX_SOCKET_PATH.Get()
	os.Remove(socketPath)
	unixSocket, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal("Failed to listen on a Unix socket.", err)
	}

	defer os.Remove(socketPath)
	// sigc := make(chan os.Signal, 1)
	// signal.Notify(sigc, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	// go func(c chan os.Signal) {
	// 	// Wait for a SIGINT or SIGKILL:
	// 	sig := <-c
	// 	fmt.Println("Caught signal %s: shutting down.", sig)
	// 	// Stop listening (and unlink the socket if unix type):
	// 	unixSocket.Close()
	// 	os.Remove(socketPath)
	// 	// And we're done:
	// 	os.Exit(0)
	// }(sigc)

	log.Info("Unix Server Started", log.NewAttr("unix_socket", socketPath))

	for {
		connection, err := unixSocket.Accept()
		if err != nil {
			log.Error("Failed to accept a unix connection.", err)
		}

		go func() {
			err := handleUnixSocketConnection(connection)
			if err != nil {
				log.Error("Error handling Unix socket connection.", err)
			}
			connection.Close()
		}()
	}
}

func handleUnixSocketConnection(conn net.Conn) error{
	var port = config.WEB_PORT.Get()
	var bufferBytes = config.BUFFER_SIZE.Get()
	var bitSize = config.BIT_SIZE.Get()

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
