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

	log.Trace("Expected to read the following bytes from connection.", log.NewAttr("expected bytes", size))
	log.Trace("Actually read the following bytes from connection.", log.NewAttr("actual bytes", numBytesRead))
	// TEST
	// fmt.Fprintln(os.Stderr, "\n\nTEST - READ")
	// fmt.Fprintln(os.Stderr, "Expected size    ", size)
	// fmt.Fprintln(os.Stderr, "Actual size    ", numBytesRead)
	// fmt.Fprintln(os.Stderr, "    ", len(buffer))
	// fmt.Fprintln(os.Stderr, "    ", buffer)
	// fmt.Fprintln(os.Stderr, "----\n")

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

	// TEST
	// fmt.Fprintln(os.Stderr, "\n\nTEST - Write")
	// fmt.Fprintln(os.Stderr, "Expected size    ", size)
	// fmt.Fprintln(os.Stderr, "    ", len(data))
	// fmt.Fprintln(os.Stderr, "    ", data)
	// fmt.Fprintln(os.Stderr, "----\n")

	err := binary.Write(responseBuffer, binary.BigEndian, size)
	if err != nil {
		return err
	}

	responseBuffer.Write(data)

	// fmt.Fprintln(os.Stderr, "starting to write...")
	log.Trace("Expected to write the following bytes from connection.", log.NewAttr("expected bytes", size))

	numBytesWritten, err := connection.Write(responseBuffer.Bytes())

	log.Trace("Actually wrote the following bytes from connection.", log.NewAttr("actual bytes", numBytesWritten))
	// TEST
	// fmt.Fprintln(os.Stderr, "Actual size    ", numBytesWritten)
	// fmt.Fprintln(os.Stderr, "    err:", err)

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
