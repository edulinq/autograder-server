package upsert

import (
	"path/filepath"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
)

func TestFileSpec(test *testing.T) {
	defer db.ResetForTesting()

	testdataDir := config.GetTestdataDir()

	emptyDir := util.MustMkDirTemp("test-internal.api.courses.upsert.filespec-")

	testCases := []struct {
		email                string
		path                 string
		expectedLocator      string
		expectedCount        int
		expectedSuccessCount int
	}{
		{"server-creator", filepath.Join(testdataDir, "course101"), "", 1, 1},
		{"server-creator", testdataDir, "", 5, 5},
		{"server-creator", emptyDir, "", 0, 0},

		{"server-creator", "", "-614", 0, 0},

		{"server-user", emptyDir, "-041", 0, 0},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		fields := map[string]any{
			"filespec": map[string]any{
				"type": common.FILESPEC_TYPE_PATH,
				"path": testCase.path,
			},
		}

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`courses/upsert/filespec`), fields, nil, testCase.email)

		if !response.Success {
			if testCase.expectedLocator != "" {
				if testCase.expectedLocator != response.Locator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
						i, testCase.expectedLocator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		if testCase.expectedLocator != "" {
			test.Errorf("Case %d: Did not get an expected error.", i)
			continue
		}

		var responseContent FileSpecResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		actualCount := len(responseContent.Results)
		if testCase.expectedCount != actualCount {
			test.Errorf("Case %d: Unexpected course count. Expected: %d, actual: %d.",
				i, testCase.expectedCount, actualCount)
			continue
		}

		actualSuccessCount := 0
		for _, result := range responseContent.Results {
			if result.Success {
				actualSuccessCount++
			}
		}

		if testCase.expectedSuccessCount != actualSuccessCount {
			test.Errorf("Case %d: Unexpected successful course count. Expected: %d, actual: %d.",
				i, testCase.expectedSuccessCount, actualSuccessCount)
			continue
		}
	}
}
