package util

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	// TEST
	"os"
)

const (
	MAX_SOCKET_MESSAGE_SIZE_BYTES = 2 * 1024 * 1024 * 1024 // 2 GB
	RESPONSE_BUFFER_SIZE          = 8
)

// Read a message from a network connection.
// The first 8 bytes is the size of the message content in bytes.
// The remaining bytes should be x bytes of the actual message content.
func ReadFromNetworkConnection(connection net.Conn) ([]byte, error) {
	sizeBuffer := make([]byte, RESPONSE_BUFFER_SIZE)

	_, err := connection.Read(sizeBuffer)
	if err != nil {
		return nil, err
	}

	size := binary.BigEndian.Uint64(sizeBuffer)
	if size > MAX_SOCKET_MESSAGE_SIZE_BYTES {
		return nil, fmt.Errorf("Message content is too large to read.")
	}

	jsonBuffer := make([]byte, size)
	numBytesRead, err := connection.Read(jsonBuffer)
	if err != nil {
		return nil, err
	}

	// TEST
	fmt.Fprintln(os.Stderr, "\n\nTEST - READ")
	fmt.Fprintln(os.Stderr, "Expected size    ", size)
	fmt.Fprintln(os.Stderr, "Actual size    ", numBytesRead)
	fmt.Fprintln(os.Stderr, "    ", len(jsonBuffer))
	// fmt.Fprintln(os.Stderr, "    ", jsonBuffer)
	fmt.Fprintln(os.Stderr, "----\n")

	return jsonBuffer, nil
}

// Write a message to a network connection.
// The first 8 bytes is the size of the message content in bytes.
// The remaining bytes should be x bytes of the actual message content.
func WriteToNetworkConnection(connection net.Conn, data []byte) error {
	size := uint64(len(data))

	if size > MAX_SOCKET_MESSAGE_SIZE_BYTES {
		return fmt.Errorf("Message content is too large to write.")
	}

	responseBuffer := new(bytes.Buffer)

	// TEST
	fmt.Fprintln(os.Stderr, "\n\nTEST - Write")
	fmt.Fprintln(os.Stderr, "Expected size    ", size)
	fmt.Fprintln(os.Stderr, "    ", len(data))
	// fmt.Fprintln(os.Stderr, "    ", data)
	fmt.Fprintln(os.Stderr, "----\n")

	err := binary.Write(responseBuffer, binary.BigEndian, size)
	if err != nil {
		fmt.Fprintln(os.Stderr, "binary write err:", err)
		return err
	}

	responseBuffer.Write(data)

	fmt.Fprintln(os.Stderr, "starting to write...")
	numBytesWritten, err := connection.Write(responseBuffer.Bytes())
	// TEST
	fmt.Fprintln(os.Stderr, "Actual size    ", numBytesWritten)
	fmt.Fprintln(os.Stderr, "    err:", err)

	if err != nil {
		return err
	}

	return nil
}

func GetUnusedPort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	listener.Close()

	return listener.Addr().(*net.TCPAddr).Port, nil
}
