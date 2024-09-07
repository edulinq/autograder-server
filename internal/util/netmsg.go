package util

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

const (
	MAX_FORM_MEM_SIZE_BYTES = 10 << 20 // 20 MB
	RESPONSE_BUFFER_SIZE    = 8
)

// Read a message from a network connection.
// First 8 bytes: the size of the message being read.
// Remaining bytes: the message content.
func ReadFromNetworkConnection(connection net.Conn) ([]byte, error) {
	sizeBuffer := make([]byte, RESPONSE_BUFFER_SIZE)

	_, err := connection.Read(sizeBuffer)
	if err != nil {
		return nil, err
	}

	size := binary.BigEndian.Uint64(sizeBuffer)
	jsonBuffer := make([]byte, size)

	_, err = connection.Read(jsonBuffer)
	if err != nil {
		return nil, err
	}

	return jsonBuffer, nil
}

// Write a message to a network connection.
// First 8 bytes: the size of the message being written.
// Remaining bytes: the message content.
func WriteToNetworkConnection(connection net.Conn, data []byte) error {
	size := uint64(len(data))

	if size > MAX_FORM_MEM_SIZE_BYTES {
		return fmt.Errorf("Payload size is too large.")
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
