package util

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/edulinq/autograder/internal/log"
)

func ReadFile(path string) (string, error) {
	data, err := ReadBinaryFile(path)
	if err != nil {
		return "", err
	}

	return string(data[:]), nil
}

func ReadBinaryFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file '%s': '%w'.", path, err)
	}

	return data, nil
}

func WriteBinaryFile(data []byte, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Failed to create file '%s': '%w'.", path, err)
	}
	defer file.Close()

	count, err := file.Write(data)
	if err != nil {
		return fmt.Errorf("Failed to write contents of file: '%w'.", err)
	}

	if len(data) != count {
		return fmt.Errorf("Write length not as expected. Expected: %d, Actual: %d.", len(data), count)
	}

	err = file.Sync()
	if err != nil {
		return fmt.Errorf("Failed to sync file: '%w'.", err)
	}

	return nil
}

func WriteFile(text string, path string) error {
	return WriteBinaryFile([]byte(text), path)
}

func WriteFileFromReader(path string, reader io.Reader) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Failed to create file '%s': '%w'.", path, err)
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("Failed to copy contents of reader: '%w'.", err)
	}

	err = file.Sync()
	if err != nil {
		return fmt.Errorf("Failed to sync file: '%w'.", err)
	}

	return nil
}

// Read a separated file into a slice of slices.
func ReadSeparatedFile(path string, delim string, skipRows int) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to open file '%s': '%w'.", path, err)
	}
	defer file.Close()

	rows := make([][]string, 0)

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		if skipRows > 0 {
			skipRows--
			continue
		}

		rows = append(rows, strings.Split(scanner.Text(), delim))
	}

	return rows, nil
}

func MustCreateEmptyFile(path string) {
	file, err := os.Create(path)
	if err != nil {
		log.Fatal("Unable to create empty file.", err, log.NewAttr("path", path))
	}
	defer file.Close()

	err = file.Sync()
	if err != nil {
		log.Fatal("Unable to sync empty file.", err, log.NewAttr("path", path))
	}
}

func MustReadFile(path string) string {
	data, err := ReadFile(path)
	if err != nil {
		log.Fatal("Unable to read file.", err, log.NewAttr("path", path))
	}

	return data
}
