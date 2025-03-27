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

	expectedSize := binary.BigEndian.Uint64(sizeBuffer)
	if expectedSize > MAX_SOCKET_MESSAGE_SIZE_BYTES {
		return nil, fmt.Errorf("Message content is too large to read.")
	}

	log.Trace("Attempting a read from a network connection.", log.NewAttr("expected-bytes", expectedSize))

	buffer := make([]byte, expectedSize)
	numBytesRead := uint64(0)

	// A connection may not be able to send all of the data in a single write.
	// Keep reading until all bytes are in the buffer.
	for numBytesRead < expectedSize {
		currentBuffer := buffer[numBytesRead:]
		currentBytesRead, err := connection.Read(currentBuffer)
		if err != nil {
			return nil, fmt.Errorf("Failed to read from the network connection. Expected to read '%d' bytes, actually read '%d' bytes: '%v'.",
				expectedSize, currentBytesRead, err)
		}

		if currentBytesRead == 0 {
			return nil, fmt.Errorf("Failed to read any bytes from the network connection. Expected to read '%d' bytes, actually read '%d' bytes for a total of '%d' bytes.",
				expectedSize, currentBytesRead, numBytesRead)
		}

		numBytesRead += uint64(currentBytesRead)
	}

	if numBytesRead > expectedSize {
		return nil, fmt.Errorf("Read too many bytes from a network connection. Expected: '%d', actual: '%d'.", expectedSize, numBytesRead)
	}

	log.Trace("Read the following bytes from a network connection.", log.NewAttr("actual-bytes", numBytesRead))

	return buffer, nil
}

// Write a message to a network connection.
// The first 8 bytes is the size of the message content in bytes.
// The remaining bytes should be x bytes of the actual message content.
func WriteToNetworkConnection(connection net.Conn, data []byte) error {
	expectedSize := uint64(len(data))

	if expectedSize > MAX_SOCKET_MESSAGE_SIZE_BYTES {
		return fmt.Errorf("Message content is too large to write.")
	}

	responseBuffer := new(bytes.Buffer)

	err := binary.Write(responseBuffer, binary.BigEndian, expectedSize)
	if err != nil {
		return err
	}

	responseBuffer.Write(data)

	log.Trace("Attempting a write to a network connection.", log.NewAttr("expected-bytes", expectedSize))

	// connection.Write() blocks until the entire buffer is written.
	numBytesWritten, err := connection.Write(responseBuffer.Bytes())
	if err != nil {
		return err
	}

	log.Trace("Wrote the following bytes to a network connection.", log.NewAttr("actual-bytes", numBytesWritten))

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
