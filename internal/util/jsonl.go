package util

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func AppendJSONLFile[T any](path string, record T) error {
	return AppendJSONLFileMany(path, []T{record})
}

func AppendJSONLFileMany[T any](path string, records []T) error {
	if !PathExists(path) {
		MkDir(filepath.Dir(path))
	}

	lines := make([]string, 0, len(records))
	for i, record := range records {
		line, err := ToJSON(record)
		if err != nil {
			return fmt.Errorf("Failed to convert record %d to JSON: '%w'.", i, err)
		}

		lines = append(lines, line)
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Failed to open JSONL file '%s': '%w'.", path, err)
	}
	defer file.Close()

	_, err = file.WriteString(strings.Join(lines, "\n") + "\n")
	if err != nil {
		return fmt.Errorf("Failed to write %d records to JSONL file '%s': '%w'.", len(lines), path, err)
	}

	err = file.Sync()
	if err != nil {
		return fmt.Errorf("Failed to sync %d records to JSONL file '%s': '%w'.", len(lines), path, err)
	}

	return nil
}

// Go through a JSONL file and call a functon for each record.
func ApplyJSONLFile[T any](path string, emptyRecord T, applyFunc func(index int, record *T, line string)) error {
	if !PathExists(path) {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Failed to open JSONL file '%s': '%w'.", path, err)
	}
	defer file.Close()

	lineno := 0
	reader := bufio.NewReader(file)
	for {
		line, err := readline(reader)
		if err != nil {
			return fmt.Errorf("Failed to read line %d from JSONL file '%s': '%w'.", lineno+1, path, err)
		}

		if line == nil {
			// EOF.
			break
		}

		lineno++

		var record *T = reflect.New(reflect.TypeOf(emptyRecord)).Interface().(*T)
		err = JSONFromBytes(line, record)
		if err != nil {
			return fmt.Errorf("Failed to convert line %d from JSONL file '%s' to JSON: '%w'.", lineno, path, err)
		}

		applyFunc(lineno-1, record, strings.TrimSpace(string(line)))
	}

	return nil
}

// Go through a JSONL file and return any records that match (matchFunc returns true).
func FilterJSONLFile[T any](path string, emptyRecord T, matchFunc func(record *T) bool) ([]*T, error) {
	records := make([]*T, 0)

	err := ApplyJSONLFile(path, emptyRecord, func(index int, record *T, line string) {
		if matchFunc(record) {
			records = append(records, record)
		}
	})

	if err != nil {
		return nil, err
	}

	return records, nil
}

// Will only return a nil content on error or EOF.
func readline(reader *bufio.Reader) ([]byte, error) {
	var isPrefix bool = true
	var err error

	var line []byte
	var fullLine []byte

	for isPrefix && err == nil {
		line, isPrefix, err = reader.ReadLine()
		fullLine = append(fullLine, line...)
	}

	if err == io.EOF {
		if fullLine == nil {
			return nil, nil
		}

		return fullLine, nil
	}

	return fullLine, err
}

// Remove any entry from the file where the function returns true.
// Because files may be large, this will be done in a streaming fashion
// where records that are not removed are written to a temp file.
// Once complete, the temp file will be moved over the actual file.
// The caller should ensure that this file is not touched by anyone else for the duration of this call.
// On most errors, the old file will remain the same.
// The exception is if there is an error on the file move operation.
func RemoveEntriesJSONLFile[T any](path string, emptyRecord T, shouldRemoveFunc func(record *T) bool) error {
	if !PathExists(path) {
		return nil
	}

	tempDir, err := MkDirTemp("jsonl-remove-entries-")
	if err != nil {
		return fmt.Errorf("Failed to create temp dir: '%w'.", err)
	}
	defer RemoveDirent(tempDir)

	tempPath := filepath.Join(tempDir, "temp.jsonl")
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("Failed to create temp file: '%w'.", err)
	}
	defer RemoveDirent(tempPath)
	defer file.Close()

	var writeError error
	applyFunc := func(index int, record *T, line string) {
		if writeError != nil {
			return
		}

		if shouldRemoveFunc(record) {
			return
		}

		_, err := file.WriteString(line + "\n")
		if err != nil {
			writeError = fmt.Errorf("Failed to write to temp file: '%w'.", err)
		}
	}

	err = ApplyJSONLFile(path, emptyRecord, applyFunc)
	if err != nil {
		return fmt.Errorf("Failed to read existing JSONL file: '%w'.", err)
	}

	if writeError != nil {
		return writeError
	}

	err = file.Sync()
	if err != nil {
		return fmt.Errorf("Failed to sync temp file: '%w'.", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("Failed to close temp file: '%w'.", err)
	}

	err = MoveDirent(tempPath, path)
	if err != nil {
		return fmt.Errorf("Failed to move temp file over existing JSONL file: '%w'.", err)
	}

	return nil
}
