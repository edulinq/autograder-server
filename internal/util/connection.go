package util

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/edulinq/autograder/internal/log"
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

	log.Trace("Reading the following bytes from a connection...", log.NewAttr("expected bytes", size))

	buffer := make([]byte, size)
	var numBytesRead uint64
	for numBytesRead = 0; numBytesRead < size; {
		currBuffer := buffer[numBytesRead:]
		currBytesRead, err := connection.Read(currBuffer)
		if err != nil {
			return nil, err
		}

		numBytesRead += uint64(currBytesRead)
	}

	log.Trace("Read the following bytes from a connection.", log.NewAttr("actual bytes", numBytesRead))

	return buffer, nil
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

	err := binary.Write(responseBuffer, binary.BigEndian, size)
	if err != nil {
		return err
	}

	responseBuffer.Write(data)

	log.Trace("Writing the following bytes to a connection...", log.NewAttr("expected bytes", size))

	numBytesWritten, err := connection.Write(responseBuffer.Bytes())
	log.Trace("Wrote the following bytes to a connection.", log.NewAttr("actual bytes", numBytesWritten))
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
