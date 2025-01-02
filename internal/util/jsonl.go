package util

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"reflect"
)

func FilterJSONLFile[T any](path string, emptyRecord T, matchFunc func(record *T) bool) ([]*T, error) {
	records := make([]*T, 0)

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to open JSONL file '%s': '%w'.", path, err)
	}
	defer file.Close()

	lineno := 0
	reader := bufio.NewReader(file)
	for {
		line, err := readline(reader)
		if err != nil {
			return nil, fmt.Errorf("Failed to read line from JSONL file '%s': '%w'.", path, err)
		}

		if line == nil {
			// EOF.
			break
		}

		lineno++

		var record *T = reflect.New(reflect.TypeOf(emptyRecord)).Interface().(*T)
		err = JSONFromBytes(line, record)
		if err != nil {
			return nil, fmt.Errorf("Failed to convert line %d from JSONL file '%s' to JSON: '%w'.", lineno, path, err)
		}

		if !matchFunc(record) {
			continue
		}

		records = append(records, record)
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
