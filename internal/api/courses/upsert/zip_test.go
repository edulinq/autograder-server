package upsert

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
)

func TestZipFile(test *testing.T) {
	defer db.ResetForTesting()

	testdataDir := config.GetTestdataDir()
	course101Dir := filepath.Join(testdataDir, "course101")

	emptyDir := util.MustMkDirTemp("test-internal.api.courses.upsert.zip-empty-")

	testCases := []commonTestCase{
		{"server-creator", course101Dir, "", 1, 1},
		{"server-creator", testdataDir, "", 2, 2},
		{"server-creator", emptyDir, "", 0, 0},

		{"server-user", course101Dir, "-041", 0, 0},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		tempDir := util.MustMkDirTemp("test-internal.api.courses.upsert.zip-prep-")
		tempPath := filepath.Join(tempDir, "test.zip")

		err := util.Zip(testCase.path, tempPath, true)
		if err != nil {
			test.Errorf("Case %d: Failed to zip source data: '%v'.", i, err)
			continue
		}

		fields := map[string]any{}

		response := core.SendTestAPIRequestFull(test, core.makeFullAPIPath(`courses/upsert/zip`), fields, []string{tempPath}, testCase.email)

		processRsponse(test, response, testCase, fmt.Sprintf("Case %d: ", i))
	}
}
