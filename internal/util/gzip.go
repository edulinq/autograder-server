package util

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func GzipFileToBytes(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Could not open source file for gzip '%s': '%w'.", path, err)
	}
	defer file.Close()

	return ReaderToGzipBytesFull(file, filepath.Base(path), "")
}

func ReaderToGzipBytes(reader io.Reader) ([]byte, error) {
	return ReaderToGzipBytesFull(reader, "", "")
}

// Read bytes from a reader and convert them to buffer of Gzipped bytes.
// Will not close the passed in reader.
func ReaderToGzipBytesFull(reader io.Reader, name string, comment string) ([]byte, error) {
	var buffer bytes.Buffer

	writer := gzip.NewWriter(&buffer)
	writer.Name = name
	writer.Comment = comment

	_, err := io.Copy(writer, reader)
	if err != nil {
		return nil, fmt.Errorf("Could not copy reader into gzip writer: '%w'.", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("Failed to close gzip writer: '%w'.", err)
	}

	return buffer.Bytes(), nil
}

func GzipBytesToFile(data []byte, path string) error {
	reader, err := gzip.NewReader(bytes.NewBuffer(bytes.Clone(data)))
	if err != nil {
		return fmt.Errorf("Failed to create gzip read for data to go in '%s': '%w'.", path, err)
	}

	clearData, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("Failed to read gzip contents to go in '%s': '%w'.", path, err)
	}

	return WriteBinaryFile(clearData, path)
}

// Gzip each file in a directory to bytes and return the output as a map: {<relpath>: bytes, ...}.
// Complements GzipBytesToDirectory().
func GzipDirectoryToBytes(baseDir string) (map[string][]byte, error) {
	fileContents := make(map[string][]byte)

	paths, err := FindFiles("", baseDir)
	if err != nil {
		return nil, fmt.Errorf("Unable to find files in base dir '%s': '%w'.", baseDir, err)
	}

	for _, path := range paths {
		contents, err := GzipFileToBytes(path)
		if err != nil {
			return nil, fmt.Errorf("Failed to gzip file '%s': '%w'.", path, err)
		}

		relPath := RelPath(path, baseDir)
		if relPath == "" {
			relPath = filepath.Base(path)
		}

		fileContents[relPath] = contents
	}

	return fileContents, nil
}

// Writes all the gzipped files into the provided dir.
// Complements GzipDirectoryToBytes().
func GzipBytesToDirectory(baseDir string, fileContents map[string][]byte) error {
	for relPath, data := range fileContents {
		path := filepath.Join(baseDir, relPath)

		err := MkDir(filepath.Dir(path))
		if err != nil {
			return err
		}

		err = GzipBytesToFile(data, path)
		if err != nil {
			return err
		}
	}

	return nil
}
