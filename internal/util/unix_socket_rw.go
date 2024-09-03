package util

import (
	"bytes"
	"encoding/binary"
	"net"
)

const RESPONSE_BUFFER_SIZE = 8

func ReadFromUnixSocket(connection net.Conn) ([]byte, error) {
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

func WriteToUnixSocket(connection net.Conn, data []byte) error {
	size := uint64(len(data))
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
