package util

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestGetDescriptionFromFunctionBase(test *testing.T) {
	tempDir := MustMkDirTemp("test-description-parser-")
	path := filepath.Join(tempDir, "test_file.go")

	content := `package temp

// Super helpful function comment!
func FunctionWithComment(){}`

	err := WriteFile(content, path)
	if err != nil {
		test.Fatalf("Failed to write function with comment: '%v'.", err)
	}

	re := regexp.MustCompile(`^FunctionWithComment$`)
	description, err := GetDescriptionFromFunction(path, re)
	if err != nil {
		test.Fatalf("Error while getting a description from a function: '%v'.", err)
	}

	expected := "Super helpful function comment!"
	if strings.Compare(description, expected) != 0 {
		test.Fatalf("Unexpected function description. Expected: '%s', actual: '%s'.", expected, description)
	}
}

func TestGetDescriptionFromFunctionMissing(test *testing.T) {
	tempDir := MustMkDirTemp("test-description-parser-")
	path := filepath.Join(tempDir, "test_file.go")

	content := `package temp

func FunctionWithoutComment(){}`

	err := WriteFile(content, path)
	if err != nil {
		test.Fatalf("Failed to write function without comment: '%v'.", err)
	}

	re := regexp.MustCompile(`^FunctionWithoutComment$`)
	_, err = GetDescriptionFromFunction(path, re)
	if err == nil {
		test.Fatalf("Missing description did not return an error.")
	}
}

func TestGetDescriptionFromFunctionWhitespace(test *testing.T) {
	tempDir := MustMkDirTemp("test-description-parser-")
	path := filepath.Join(tempDir, "test_file.go")

	content := `package temp

//    	  	
func FunctionWithWhitespaceComment(){}`

	err := WriteFile(content, path)
	if err != nil {
		test.Fatalf("Failed to write function with whitespace comment: '%v'.", err)
	}

	re := regexp.MustCompile(`^FunctionWithWhitespaceComment$`)
	description, err := GetDescriptionFromFunction(path, re)
	if err != nil {
		test.Fatalf("Error while getting a description from a function: '%v'.", err)
	}

	expected := ""
	if strings.Compare(description, expected) != 0 {
		test.Fatalf("Unexpected function description. Expected: '%s', actual: '%s'.", expected, description)
	}
}
