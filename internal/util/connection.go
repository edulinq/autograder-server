package util

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

const (
	MAX_SOCKET_MESSAGE_SIZE_BYTES = 2 << 30 // 2 GB
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
	_, err = connection.Read(jsonBuffer)
	if err != nil {
		return nil, err
	}

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

	err := binary.Write(responseBuffer, binary.BigEndian, size)
	if err != nil {
		return err
	}

	responseBuffer.Write(data)

	_, err = connection.Write(responseBuffer.Bytes())
	if err != nil {
		return err
	}

	return nil
}
