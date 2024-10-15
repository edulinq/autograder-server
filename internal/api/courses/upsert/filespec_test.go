package upsert

import (
	"fmt"
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

	testCases := []commonTestCase{
		{"server-creator", filepath.Join(testdataDir, "course101"), "", 1, 1},
		{"server-creator", testdataDir, "", 2, 2},
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

		processRsponse(test, response, testCase, fmt.Sprintf("Case %d: ", i))
	}
}
