package util

import (
	"bufio"
	"fmt"
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
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("Failed to write file '%s': '%w'.", path, err)
	}

	return nil
}

func WriteFile(text string, path string) error {
	return WriteBinaryFile([]byte(text), path)
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

func MustCreateFile(path string) *os.File {
	file, err := os.Create(path)
	if err != nil {
		log.Fatal("Unable to create file.", err, log.NewAttr("path", path))
	}

	return file
}

func MustReadFile(path string) string {
	data, err := ReadFile(path)
	if err != nil {
		log.Fatal("Unable to read file.", err, log.NewAttr("path", path))
	}

	return data
}
