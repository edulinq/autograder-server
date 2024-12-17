package util

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestGetDescriptionFromFunction(test *testing.T) {
	tempDir := MustMkDirTemp("test-description-parser-")
	path := filepath.Join(tempDir, "test_file.go")

	testCases := []struct {
		pattern             string
		content             string
		expectedDescription string
		descriptionErr      bool
	}{
		// Valid descriptions.
		{`^FunctionWithComment$`, base_description, "Super helpful function comment!", false},
		{`^FunctionWithWhitespaceComment$`, whitespace_description, "", false},

		// Invalid descriptions.
		{`^FunctionWithoutComment$`, missing_description, "", true},
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
			if testCase.descriptionErr {
				if !strings.Contains(err.Error(), "Unable to find a description") {
					test.Errorf("Case %d: Incorrect error returned when getting function description: '%v'.", i, err)
				}
			} else {
				test.Errorf("Case %d: Error while getting a description from a function: '%v'.", i, err)
			}

			continue
		}

		if testCase.descriptionErr {
			test.Errorf("Case %d: Did not get an expected description error.", i)
			continue
		}

		if description != testCase.expectedDescription {
			test.Errorf("Case %d: Unexpected function description. Expected: '%s', actual: '%s'.", i, testCase.expectedDescription, description)
			continue
		}
	}
}

const base_description = `package temp

// Super helpful function comment!
func FunctionWithComment(){}`

const missing_description = `package temp

func FunctionWithoutComment(){}`

const whitespace_description = `package temp

//    	  	
func FunctionWithWhitespaceComment(){}`
