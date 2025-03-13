package common

import (
	"strings"
	"testing"
)

func TestValidateIDBase(test *testing.T) {
	testCases := []struct {
		input          string
		expected       string
		errorSubstring string
	}{
		// Valid
		{"a", "a", ""},
		{"A", "a", ""},
		{"a.a", "a.a", ""},
		{"a-a", "a-a", ""},
		{"a_a", "a_a", ""},

		// Errors

		// Empty
		{"", "", "not be empty"},

		// Bad Characters
		{"a!", "", "must only have"},
		{"a a", "", "must only have"},

		// Bad Terminals
		{".a", "", "cannot start or end with"},
		{"a.", "", "cannot start or end with"},
		{"-a", "", "cannot start or end with"},
		{"a-", "", "cannot start or end with"},
		{"_a", "", "cannot start or end with"},
		{"a_", "", "cannot start or end with"},
	}

	for i, testCase := range testCases {
		actual, err := ValidateID(testCase.input)
		if err != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error outpout. Expected Substring '%s', Actual Error: '%v'.", i, testCase.errorSubstring, err)
				}
			} else {
				test.Errorf("Case %d: Failed to validate id '%s': '%v'.", i, testCase.input, err)
			}

			continue
		}

		if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected error.", i)
			continue
		}

		if testCase.expected != actual {
			test.Errorf("Case %d: Validation not as expected. Expected: '%+v', Actual: '%+v'.", i, testCase.expected, actual)
			continue
		}

		doubleClean, err := ValidateID(actual)
		if err != nil {
			test.Errorf("Case %d: Failed to re-validate id '%s': '%v'.", i, actual, err)
			continue
		}

		if testCase.expected != doubleClean {
			test.Errorf("Case %d: Re-vlidation not as expected. Expected: '%+v', Actual: '%+v'.", i, testCase.expected, doubleClean)
			continue
		}
	}
}
