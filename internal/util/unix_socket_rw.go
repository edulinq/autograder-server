package util

import (
	"bytes"
	"encoding/binary"
	"net"
)

func ReadFromUnixSocket(conn net.Conn, bufferBytes int) ([]byte, error) {
	sizeBuffer := make([]byte, bufferBytes)

	_, err := conn.Read(sizeBuffer)
	if err != nil {
		return nil, err
	}

	size := binary.BigEndian.Uint64(sizeBuffer)

	jsonBuffer := make([]byte, size)
	_, err = conn.Read(jsonBuffer)
	if err != nil {
		return nil, err
	}
	return jsonBuffer, nil
}

func WriteToUnixSocket(conn net.Conn, data []byte) error {
	size := uint64(len(data))
	responseBuffer := new(bytes.Buffer)

	err := binary.Write(responseBuffer, binary.BigEndian, size)
	if err != nil {
		return err
	}

	responseBuffer.Write(data)

	_, err = conn.Write(responseBuffer.Bytes())
	if err != nil {
		return err
	}

	return nil
}
