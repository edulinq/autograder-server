package util

import (
	"path/filepath"
	"regexp"
	"testing"
)

func TestGetDescriptionFromFunction(test *testing.T) {
	tempDir := MustMkDirTemp("test-description-parser-")
	defer RemoveDirent(tempDir)

	path := filepath.Join(tempDir, "test_file.go")

	testCases := []struct {
		pattern             string
		content             string
		expectedDescription string
	}{
		// Valid descriptions.
		{`^FunctionWithComment$`, BASE_DESCRIPTION, "Super helpful function comment!"},
		{`^FunctionWithoutComment$`, MISSING_DESCRIPTION, ""},
		{`^FunctionWithWhitespaceComment$`, WHITESPACE_DESCRIPTION, ""},
	}

	for i, testCase := range testCases {
		err := WriteFile(testCase.content, path)
		if err != nil {
			test.Errorf("Case %d: Failed to write content to the test file: '%v'.", i, err)
			continue
		}

		functionNamePattern := regexp.MustCompile(testCase.pattern)
		description, err := GetDescriptionFromFunction(path, functionNamePattern)
		if err != nil {
			test.Errorf("Case %d: Error while getting a description from a function: '%v'.", i, err)
			continue
		}

		if description != testCase.expectedDescription {
			test.Errorf("Case %d: Unexpected function description. Expected: '%s', actual: '%s'.", i, testCase.expectedDescription, description)
			continue
		}
	}
}

const BASE_DESCRIPTION = `package temp

// Super helpful function comment!
func FunctionWithComment(){}`

const MISSING_DESCRIPTION = `package temp

func FunctionWithoutComment(){}`

const WHITESPACE_DESCRIPTION = `package temp

//    	  	
func FunctionWithWhitespaceComment(){}`
