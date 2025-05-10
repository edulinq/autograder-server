package db

import (
	"reflect"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

func TestParseServerUserReferenceBase(test *testing.T) {
	testCases := []struct {
		reference      string
		output         *UserReference
		errorSubstring string
	}{
		{},
	}

	for i, testCase := range testCases {
		result, err := ParseUserReference(testCase.reference)
		if err != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error output. Expected Substring '%s', Actual Error: '%s'.",
						i, testCase.errorSubstring, err.Error())
				}
			} else {
				test.Errorf("Case %d: Failed to parse user reference '%s': '%v'.", i, testCase.reference, err.Error())
			}

			continue
		}

		if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected error for reference '%s'.", i, testCase.reference)
			continue
		}
	}
}
